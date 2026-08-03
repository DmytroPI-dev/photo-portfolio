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
	PresignUpload(ctx context.Context, key, contentType string, maxBytes int64) (PresignedUpload, error)
	PresignGet(ctx context.Context, key string) (string, error)
}

// PresignedUpload contains the URL and form fields required for a constrained
// S3 POST upload. The policy binds both the object key and size limit.
type PresignedUpload struct {
	URL    string
	Fields map[string]string
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

func (store *OriginalStore) PresignUpload(ctx context.Context, key, contentType string, maxBytes int64) (PresignedUpload, error) {
	request, err := store.presigner.PresignPostObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(store.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(options *s3.PresignPostOptions) {
		options.Expires = 15 * time.Minute
		options.Conditions = []any{[]any{"content-length-range", 1, maxBytes}}
	})
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("presign original upload: %w", err)
	}
	return PresignedUpload{URL: request.URL, Fields: request.Values}, nil
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
