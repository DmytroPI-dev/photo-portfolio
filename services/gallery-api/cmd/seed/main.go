// Command seed writes the deterministic placeholder portfolio to DynamoDB.
//
// It is deliberately separate from the public Lambda: the Lambda has read-only
// permissions, while this command is run locally with a human AWS identity
// during bootstrap. Re-running it overwrites only the fixed placeholder and
// canonical seed keys; never use it after real administrator edits unless a
// deliberate reset is intended.
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	table := strings.TrimSpace(os.Getenv("GALLERY_METADATA_TABLE"))
	if table == "" {
		log.Fatal("GALLERY_METADATA_TABLE must name the target DynamoDB table")
	}

	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load AWS SDK configuration: %v", err)
	}

	writes, err := gallery.SeedDynamo(ctx, dynamodb.NewFromConfig(awsConfig), table)
	if err != nil {
		log.Fatalf("seed DynamoDB table %q: %v", table, err)
	}

	log.Printf("seeded %d DynamoDB metadata records into %s", writes, table)
}
