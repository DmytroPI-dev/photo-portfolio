package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/processing"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/storage"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	worker, err := newWorker(context.Background())
	if err != nil {
		log.Fatalf("configure image worker: %v", err)
	}
	lambda.Start(worker.HandleSQSEvent)
}

func newWorker(ctx context.Context) (*processing.Worker, error) {
	table := requiredEnvironment("GALLERY_METADATA_TABLE")
	originalsBucket := requiredEnvironment("GALLERY_ORIGINALS_BUCKET")
	derivativesBucket := requiredEnvironment("GALLERY_DERIVATIVES_BUCKET")
	if table == "" || originalsBucket == "" || derivativesBucket == "" {
		return nil, fmt.Errorf("GALLERY_METADATA_TABLE, GALLERY_ORIGINALS_BUCKET, and GALLERY_DERIVATIVES_BUCKET are required")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS SDK configuration: %w", err)
	}
	repository := gallery.NewDynamoRepository(dynamodb.NewFromConfig(awsConfig), table)
	return processing.NewWorker(repository, storage.NewProcessingStore(awsConfig), processing.NewVipsTransformer(), originalsBucket, derivativesBucket), nil
}

func requiredEnvironment(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}
