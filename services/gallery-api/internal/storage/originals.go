// Package storage contains the AWS boundary for original image objects. The
// HTTP layer depends only on the small Presigner interface, which keeps the
// upload contract easy to test without AWS credentials.
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Presigner creates short-lived URLs for private original objects. The browser
// sends image bytes directly to S3; Lambda never proxies large uploads.
type Presigner interface {
	PresignPut(ctx context.Context, key, contentType string) (string, error)
	PresignGet(ctx context.Context, key string) (string, error)
}

type OriginalStore struct {
	bucket    string
	presigner *s3.PresignClient
}

func NewOriginalStore(awsConfig aws.Config, bucket string) *OriginalStore {
	return &OriginalStore{
		bucket:    bucket,
		presigner: s3.NewPresignClient(s3.NewFromConfig(awsConfig)),
	}
}

func (store *OriginalStore) PresignPut(ctx context.Context, key, contentType string) (string, error) {
	request, err := store.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(store.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("presign original upload: %w", err)
	}
	return request.URL, nil
}

func (store *OriginalStore) PresignGet(ctx context.Context, key string) (string, error) {
	request, err := store.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(10*time.Minute))
	if err != nil {
		return "", fmt.Errorf("presign original preview: %w", err)
	}
	return request.URL, nil
}
