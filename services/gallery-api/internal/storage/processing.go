package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ProcessingStore is the worker's narrow S3 boundary. It streams uploads and
// downloads through files in Lambda's ephemeral storage, avoiding unnecessary
// duplicate image buffers in Go's heap.
type ProcessingStore struct {
	client *s3.Client
}

func NewProcessingStore(awsConfig aws.Config) *ProcessingStore {
	return &ProcessingStore{client: s3.NewFromConfig(awsConfig)}
}

func (store *ProcessingStore) Download(ctx context.Context, bucket, key, destination string) error {
	object, err := store.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get s3://%s/%s: %w", bucket, key, err)
	}
	defer object.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create %q: %w", destination, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, object.Body); err != nil {
		return fmt.Errorf("write %q: %w", destination, err)
	}
	return nil
}

func (store *ProcessingStore) Upload(ctx context.Context, bucket, key, sourcePath, contentType string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open %q: %w", sourcePath, err)
	}
	defer file.Close()

	_, err = store.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         file,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("put s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}

func (store *ProcessingStore) Tag(ctx context.Context, bucket, key string, tags map[string]string) error {
	tagSet := make([]types.Tag, 0, len(tags))
	for tagKey, value := range tags {
		tagSet = append(tagSet, types.Tag{Key: aws.String(tagKey), Value: aws.String(value)})
	}
	_, err := store.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		Tagging: &types.Tagging{TagSet: tagSet},
	})
	if err != nil {
		return fmt.Errorf("tag s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}
