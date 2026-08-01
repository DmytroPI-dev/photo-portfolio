package appconfig

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
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
