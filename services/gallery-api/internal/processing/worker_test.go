package processing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/aws/aws-lambda-go/events"
)

func TestWorkerRendersResponsiveDerivativesAndMarksSourceComplete(t *testing.T) {
	photo := testPhoto()
	repository := &photoRepository{photos: map[string]gallery.AdminPhoto{photo.ID: photo}}
	objects := &objectStore{}
	transformer := transformer{size: ImageSize{Width: 1200, Height: 1800}}
	worker := NewWorker(repository, objects, transformer, "originals", "derivatives")
	worker.clock = func() time.Time { return time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC) }

	if err := worker.ProcessObject(context.Background(), photo.OriginalKey); err != nil {
		t.Fatalf("ProcessObject returned error: %v", err)
	}

	actual := repository.photos[photo.ID]
	if actual.ProcessingStatus != gallery.ProcessingReady || actual.ProcessingError != "" {
		t.Fatalf("processing result = %q / %q, want ready with no error", actual.ProcessingStatus, actual.ProcessingError)
	}
	if actual.DerivativeKey != "derivatives/photo-001/v1/large.webp" {
		t.Fatalf("large derivative key = %q", actual.DerivativeKey)
	}
	if actual.Width != 1200 || actual.Height != 1800 || actual.Version != 4 {
		t.Fatalf("updated photo = %#v", actual)
	}
	if got, want := strings.Join(objects.uploaded, ","), "derivatives/photo-001/v1/thumbnail.webp,derivatives/photo-001/v1/medium.webp,derivatives/photo-001/v1/large.webp"; got != want {
		t.Fatalf("uploaded = %q, want %q", got, want)
	}
	if got := objects.tags[photo.OriginalKey]["gallery-processing"]; got != "complete" {
		t.Fatalf("source tag = %q, want complete", got)
	}
}

func TestWorkerFailureRemainsRetryableAndRecordsTheReason(t *testing.T) {
	photo := testPhoto()
	repository := &photoRepository{photos: map[string]gallery.AdminPhoto{photo.ID: photo}}
	worker := NewWorker(repository, &objectStore{}, transformer{err: errors.New("invalid JPEG")}, "originals", "derivatives")

	err := worker.ProcessObject(context.Background(), photo.OriginalKey)
	if err == nil || !strings.Contains(err.Error(), "invalid JPEG") {
		t.Fatalf("ProcessObject error = %v, want invalid JPEG", err)
	}
	actual := repository.photos[photo.ID]
	if actual.ProcessingStatus != gallery.ProcessingFailed || !strings.Contains(actual.ProcessingError, "invalid JPEG") {
		t.Fatalf("failed photo = %#v", actual)
	}
}

func TestWorkerDuplicateReadyDeliveryOnlyEnsuresLifecycleTag(t *testing.T) {
	photo := testPhoto()
	photo.ProcessingStatus = gallery.ProcessingReady
	photo.DerivativeKey = derivativeKey(photo.ID, "large")
	repository := &photoRepository{photos: map[string]gallery.AdminPhoto{photo.ID: photo}}
	objects := &objectStore{}
	transformer := transformer{err: errors.New("must not render")}
	worker := NewWorker(repository, objects, transformer, "originals", "derivatives")

	if err := worker.ProcessObject(context.Background(), photo.OriginalKey); err != nil {
		t.Fatalf("ProcessObject duplicate returned error: %v", err)
	}
	if len(objects.uploaded) != 0 || repository.updates != 0 {
		t.Fatalf("duplicate performed work: uploads=%v updates=%d", objects.uploaded, repository.updates)
	}
	if got := objects.tags[photo.OriginalKey]["gallery-processing"]; got != "complete" {
		t.Fatalf("source tag = %q, want complete", got)
	}
}

func TestPhotoIDFromOriginalKey(t *testing.T) {
	photoID, err := photoIDFromOriginalKey("originals/photo-001/original.jpeg")
	if err != nil || photoID != "photo-001" {
		t.Fatalf("photoIDFromOriginalKey = %q, %v", photoID, err)
	}
	if _, err := photoIDFromOriginalKey("derivatives/photo-001/large.webp"); err == nil {
		t.Fatal("photoIDFromOriginalKey accepted a derivative key")
	}
}

func TestWorkerReportsOnlyFailedSQSRecordsForRetry(t *testing.T) {
	photo := testPhoto()
	repository := &photoRepository{photos: map[string]gallery.AdminPhoto{photo.ID: photo}}
	worker := NewWorker(repository, &objectStore{}, transformer{size: ImageSize{Width: 1200, Height: 1800}}, "originals", "derivatives")

	response, err := worker.HandleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: "valid-message",
			Body:      `{"Records":[{"s3":{"bucket":{"name":"originals"},"object":{"key":"originals%2Fphoto-001%2Foriginal.jpeg"}}}]}`,
		},
		{MessageId: "invalid-message", Body: `not-json`},
	}})
	if err != nil {
		t.Fatalf("HandleSQSEvent returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 1 || response.BatchItemFailures[0].ItemIdentifier != "invalid-message" {
		t.Fatalf("batch failures = %#v", response.BatchItemFailures)
	}
	if actual := repository.photos[photo.ID]; actual.ProcessingStatus != gallery.ProcessingReady {
		t.Fatalf("valid SQS record did not complete: %#v", actual)
	}
}

func testPhoto() gallery.AdminPhoto {
	return gallery.AdminPhoto{
		Photo:            gallery.Photo{ID: "photo-001", Title: "Upload", CollectionID: "drawings", Width: 1, Height: 1, Order: 1},
		Status:           gallery.PublicationDraft,
		ProcessingStatus: gallery.ProcessingPending,
		OriginalKey:      "originals/photo-001/original.jpeg",
		Version:          2,
	}
}

type photoRepository struct {
	photos  map[string]gallery.AdminPhoto
	updates int
}

func (repository *photoRepository) GetAdminPhotoByID(_ context.Context, id string) (gallery.AdminPhoto, bool, error) {
	photo, found := repository.photos[id]
	return photo, found, nil
}

func (repository *photoRepository) UpdateAdminPhoto(_ context.Context, previous, next gallery.AdminPhoto) error {
	current, found := repository.photos[previous.ID]
	if !found || current.Version != previous.Version {
		return gallery.ErrVersionConflict
	}
	repository.photos[previous.ID] = next
	repository.updates++
	return nil
}

type objectStore struct {
	uploaded []string
	tags     map[string]map[string]string
}

func (store *objectStore) Download(_ context.Context, _, _, destination string) error {
	return os.WriteFile(destination, []byte("source"), 0o600)
}

func (store *objectStore) Upload(_ context.Context, _, key, sourcePath, contentType string) error {
	if contentType != "image/webp" {
		return errors.New("unexpected content type")
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return err
	}
	store.uploaded = append(store.uploaded, key)
	return nil
}

func (store *objectStore) Tag(_ context.Context, _, key string, tags map[string]string) error {
	if store.tags == nil {
		store.tags = make(map[string]map[string]string)
	}
	store.tags[key] = tags
	return nil
}

type transformer struct {
	size ImageSize
	err  error
}

func (transformer transformer) Render(_ context.Context, _, outputDirectory string, variants []Variant) (ImageSize, []RenderedVariant, error) {
	if transformer.err != nil {
		return ImageSize{}, nil, transformer.err
	}
	rendered := make([]RenderedVariant, 0, len(variants))
	for _, variant := range variants {
		outputPath := filepath.Join(outputDirectory, variant.Name+".webp")
		if err := os.WriteFile(outputPath, []byte(variant.Name), 0o600); err != nil {
			return ImageSize{}, nil, err
		}
		rendered = append(rendered, RenderedVariant{Variant: variant, Path: outputPath})
	}
	return transformer.size, rendered, nil
}
