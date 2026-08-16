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
