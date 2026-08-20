package appconfig

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/storage"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewRepository selects the local placeholder repository when no table is
// configured, and DynamoDB for the deployed Lambda. This keeps `go run` useful
// without AWS credentials while preventing production from silently falling
// back to stale in-memory data if DynamoDB is misconfigured.
func NewRepository(ctx context.Context) (gallery.Repository, error) {
	table := strings.TrimSpace(os.Getenv("GALLERY_METADATA_TABLE"))
	if table == "" {
		return gallery.NewSeedRepository(), nil
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS SDK configuration: %w", err)
	}

	return gallery.NewDynamoRepository(dynamodb.NewFromConfig(awsConfig), table), nil
}

// NewOriginalStore is nil for the local seed-only server. In AWS, the Lambda
// receives the private originals bucket name from Terraform and can mint the
// narrowly scoped, short-lived URLs used by the admin browser.
func NewOriginalStore(ctx context.Context) (storage.Presigner, error) {
	bucket := strings.TrimSpace(os.Getenv("GALLERY_ORIGINALS_BUCKET"))
	if bucket == "" {
		return nil, nil
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS SDK configuration for originals: %w", err)
	}
	return storage.NewOriginalStore(awsConfig, bucket), nil
}

// NewProcessingQueue configures explicit processing retries in AWS. It is nil
// locally, where no SQS queue is present; the handler then returns a clear
// configuration error instead of accepting a retry it cannot deliver.
func NewProcessingQueue(ctx context.Context) (storage.ProcessingQueue, error) {
	queueURL := strings.TrimSpace(os.Getenv("GALLERY_PROCESSING_QUEUE_URL"))
	if queueURL == "" {
		return nil, nil
	}
	bucket := strings.TrimSpace(os.Getenv("GALLERY_ORIGINALS_BUCKET"))
	if bucket == "" {
		return nil, fmt.Errorf("GALLERY_ORIGINALS_BUCKET is required when GALLERY_PROCESSING_QUEUE_URL is configured")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS SDK configuration for processing queue: %w", err)
	}
	return storage.NewSQSProcessingQueue(awsConfig, queueURL, bucket), nil
}

// NewPhotoAssetDeleter enables permanent cleanup for uploaded photos. Static
// seed records have no private original and do not require this dependency.
func NewPhotoAssetDeleter(ctx context.Context) (storage.PhotoAssetDeleter, error) {
	originalsBucket := strings.TrimSpace(os.Getenv("GALLERY_ORIGINALS_BUCKET"))
	if originalsBucket == "" {
		return nil, nil
	}
	derivativesBucket := strings.TrimSpace(os.Getenv("GALLERY_DERIVATIVES_BUCKET"))
	if derivativesBucket == "" {
		return nil, fmt.Errorf("GALLERY_DERIVATIVES_BUCKET is required when GALLERY_ORIGINALS_BUCKET is configured")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS SDK configuration for photo cleanup: %w", err)
	}
	return storage.NewS3PhotoAssetDeleter(
		awsConfig,
		originalsBucket,
		derivativesBucket,
		strings.TrimSpace(os.Getenv("GALLERY_MEDIA_DISTRIBUTION_ID")),
	), nil
}

// MediaBaseURL returns the public CloudFront base URL used only when a ready
// uploaded photo is published. An empty value is valid for local development
// and intentionally prevents uploaded drafts from being published.
func MediaBaseURL() (string, error) {
	raw := strings.TrimRight(strings.TrimSpace(os.Getenv("GALLERY_MEDIA_BASE_URL")), "/")
	if raw == "" {
		return "", nil
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("GALLERY_MEDIA_BASE_URL must be an HTTPS origin without a query or fragment")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}
