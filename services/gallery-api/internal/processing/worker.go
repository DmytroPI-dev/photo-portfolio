// Package processing turns a private upload into deterministic WebP variants.
// It deliberately owns only the asynchronous boundary: the HTTP API remains
// responsible for gallery edits, while this package updates the same canonical
// photo record through its existing optimistic-concurrency contract.
package processing

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
)

const (
	// ProfileVersion becomes part of the derivative key. A future quality or
	// sizing change can safely coexist with v1 until the public URL is switched.
	ProfileVersion = "v1"

	// MaxDecodedPixels caps decompression work before libvips reads a malicious
	// image into memory. It accepts a 100-megapixel original, comfortably above
	// the browser upload limit while protecting the Lambda's fixed memory size.
	MaxDecodedPixels = 100_000_000
)

var ErrPhotoNotFound = errors.New("uploaded image has no matching photo metadata yet")

// Variant is one responsive WebP rendition. Width is a maximum: libvips never
// enlarges a smaller original, avoiding invented pixels and unnecessary cost.
type Variant struct {
	Name     string
	MaxWidth int
}

var defaultVariants = []Variant{
	{Name: "thumbnail", MaxWidth: 480},
	{Name: "medium", MaxWidth: 1200},
	{Name: "large", MaxWidth: 2400},
}

// PhotoRepository is intentionally smaller than the whole admin repository.
// Both DynamoRepository and MemoryRepository already satisfy it, and tests can
// exercise processing behavior without constructing DynamoDB transactions.
type PhotoRepository interface {
	GetAdminPhotoByID(ctx context.Context, id string) (gallery.AdminPhoto, bool, error)
	UpdateAdminPhoto(ctx context.Context, previous, next gallery.AdminPhoto) error
}

// ObjectStore isolates S3 transfers and tagging from image processing. Files
// move through /tmp instead of Go byte slices so a 25 MB input plus derivatives
// does not create multiple in-memory copies in the Lambda process.
type ObjectStore interface {
	Download(ctx context.Context, bucket, key, destination string) error
	Upload(ctx context.Context, bucket, key, sourcePath, contentType string) error
	Tag(ctx context.Context, bucket, key string, tags map[string]string) error
}

// Transformer is backed by the libvips command-line tool in the container.
// Keeping it as an interface makes the delivery and DynamoDB transition tests
// independent of a developer's locally installed libvips binary.
type Transformer interface {
	Render(ctx context.Context, inputPath, outputDirectory string, variants []Variant) (ImageSize, []RenderedVariant, error)
}

type ImageSize struct {
	Width  int
	Height int
}

type RenderedVariant struct {
	Variant
	Path string
}

type Worker struct {
	repository        PhotoRepository
	objects           ObjectStore
	transformer       Transformer
	originalsBucket   string
	derivativesBucket string
	clock             func() time.Time
	variants          []Variant
}

func NewWorker(repository PhotoRepository, objects ObjectStore, transformer Transformer, originalsBucket, derivativesBucket string) *Worker {
	return &Worker{
		repository:        repository,
		objects:           objects,
		transformer:       transformer,
		originalsBucket:   originalsBucket,
		derivativesBucket: derivativesBucket,
		clock:             time.Now,
		variants:          append([]Variant(nil), defaultVariants...),
	}
}

// ProcessObject is safe to call again for the same S3 notification. Once the
// canonical record names the deterministic large derivative as ready, the
// expensive transform is skipped and the source tag is simply ensured.
func (worker *Worker) ProcessObject(ctx context.Context, originalKey string) error {
	photoID, err := photoIDFromOriginalKey(originalKey)
	if err != nil {
		return err
	}

	photo, found, err := worker.repository.GetAdminPhotoByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("get photo %q: %w", photoID, err)
	}
	if !found {
		// The browser uploads just before creating the photo record. The queue's
		// short delay normally avoids this race; returning an error retains the
		// unusual slow request for SQS retry rather than dropping it.
		return ErrPhotoNotFound
	}
	if photo.OriginalKey != originalKey {
		return fmt.Errorf("photo %q does not own original %q", photoID, originalKey)
	}

	largeKey := derivativeKey(photo.ID, "large")
	if photo.ProcessingStatus == gallery.ProcessingReady && photo.DerivativeKey == largeKey {
		return worker.tagSourceComplete(ctx, originalKey)
	}
	if photo.ProcessingStatus != gallery.ProcessingProcessing {
		processingPhoto := photo
		processingPhoto.ProcessingStatus = gallery.ProcessingProcessing
		processingPhoto.ProcessingError = ""
		processingPhoto.Version++
		processingPhoto.UpdatedAt = worker.clock().UTC().Format(time.RFC3339)
		if err := worker.repository.UpdateAdminPhoto(ctx, photo, processingPhoto); err != nil {
			return fmt.Errorf("record processing start: %w", err)
		}
		photo = processingPhoto
	}

	workingDirectory, err := newWorkingDirectory()
	if err != nil {
		return err
	}
	defer removeWorkingDirectory(workingDirectory)

	inputPath := path.Join(workingDirectory, "original")
	if err := worker.objects.Download(ctx, worker.originalsBucket, originalKey, inputPath); err != nil {
		return worker.fail(ctx, photo, fmt.Errorf("download original: %w", err))
	}

	size, rendered, err := worker.transformer.Render(ctx, inputPath, workingDirectory, worker.variants)
	if err != nil {
		return worker.fail(ctx, photo, fmt.Errorf("render WebP derivatives: %w", err))
	}
	if size.Width < 1 || size.Height < 1 {
		return worker.fail(ctx, photo, errors.New("libvips returned invalid image dimensions"))
	}

	for _, item := range rendered {
		if err := worker.objects.Upload(ctx, worker.derivativesBucket, derivativeKey(photo.ID, item.Name), item.Path, "image/webp"); err != nil {
			return worker.fail(ctx, photo, fmt.Errorf("upload %s derivative: %w", item.Name, err))
		}
	}

	next := photo
	next.ProcessingStatus = gallery.ProcessingReady
	next.ProcessingError = ""
	next.DerivativeKey = largeKey
	// Do not expose a private derivative key as Photo.Src. The API derives the
	// public CloudFront URL from DerivativeKey only during the publish transition,
	// after this worker has made every derivative available.
	next.Width = size.Width
	next.Height = size.Height
	next.Version++
	next.UpdatedAt = worker.clock().UTC().Format(time.RFC3339)
	if err := worker.repository.UpdateAdminPhoto(ctx, photo, next); err != nil {
		// A duplicate worker may have completed first. The following SQS retry
		// re-reads the record and exits through the idempotent ready branch.
		return fmt.Errorf("record processing result: %w", err)
	}

	if err := worker.tagSourceComplete(ctx, originalKey); err != nil {
		return fmt.Errorf("tag successful source: %w", err)
	}
	return nil
}

func (worker *Worker) fail(ctx context.Context, photo gallery.AdminPhoto, cause error) error {
	message := cause.Error()
	if len(message) > 1000 {
		message = message[:1000]
	}
	next := photo
	next.ProcessingStatus = gallery.ProcessingFailed
	next.ProcessingError = message
	next.Version++
	next.UpdatedAt = worker.clock().UTC().Format(time.RFC3339)
	if err := worker.repository.UpdateAdminPhoto(ctx, photo, next); err != nil {
		return fmt.Errorf("%w (record failure: %v)", cause, err)
	}
	return cause
}

func (worker *Worker) tagSourceComplete(ctx context.Context, originalKey string) error {
	return worker.objects.Tag(ctx, worker.originalsBucket, originalKey, map[string]string{"gallery-processing": "complete"})
}

func derivativeKey(photoID, variant string) string {
	return path.Join("derivatives", photoID, ProfileVersion, variant+".webp")
}

func photoIDFromOriginalKey(originalKey string) (string, error) {
	parts := strings.Split(originalKey, "/")
	if len(parts) != 3 || parts[0] != "originals" || parts[1] == "" || !validOriginalFilename(parts[2]) {
		return "", fmt.Errorf("unsupported original key %q", originalKey)
	}
	return parts[1], nil
}

func validOriginalFilename(filename string) bool {
	switch filename {
	case "original.jpg", "original.png", "original.webp":
		return true
	default:
		return false
	}
}
