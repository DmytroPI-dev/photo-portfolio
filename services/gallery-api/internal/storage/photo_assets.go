package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cloudfronttypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// PhotoAssetDeleter removes the private original and every generated
// derivative for a photo. It is intentionally separate from Presigner: only a
// permanent archived-photo deletion needs these stronger capabilities.
type PhotoAssetDeleter interface {
	DeletePhotoAssets(ctx context.Context, photoID, originalKey string) error
}

type cloudFrontInvalidator interface {
	CreateInvalidation(ctx context.Context, params *cloudfront.CreateInvalidationInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateInvalidationOutput, error)
}

// S3PhotoAssetDeleter keeps originals and derivatives in their distinct
// private buckets. When public media delivery exists, it requests an
// invalidation before removing origin objects, preventing a permanent delete
// from leaving a known derivative URL available from a long-lived CDN cache.
type S3PhotoAssetDeleter struct {
	s3Client          *s3.Client
	originalsBucket   string
	derivativesBucket string
	distributionID    string
	invalidator       cloudFrontInvalidator
}

func NewS3PhotoAssetDeleter(awsConfig aws.Config, originalsBucket, derivativesBucket, distributionID string) *S3PhotoAssetDeleter {
	deleter := &S3PhotoAssetDeleter{
		s3Client:          s3.NewFromConfig(awsConfig),
		originalsBucket:   originalsBucket,
		derivativesBucket: derivativesBucket,
		distributionID:    distributionID,
	}
	if distributionID != "" {
		deleter.invalidator = cloudfront.NewFromConfig(awsConfig)
	}
	return deleter
}

func (deleter *S3PhotoAssetDeleter) DeletePhotoAssets(ctx context.Context, photoID, originalKey string) error {
	prefix := "derivatives/" + photoID + "/"
	if deleter.invalidator != nil {
		_, err := deleter.invalidator.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(deleter.distributionID),
			InvalidationBatch: &cloudfronttypes.InvalidationBatch{
				CallerReference: aws.String(fmt.Sprintf("gallery-delete-%s-%d", photoID, time.Now().UnixNano())),
				Paths: &cloudfronttypes.Paths{
					Quantity: aws.Int32(1),
					Items:    []string{"/" + prefix + "*"},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("invalidate cached derivatives: %w", err)
		}
	}

	if originalKey != "" {
		if _, err := deleter.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(deleter.originalsBucket),
			Key:    aws.String(originalKey),
		}); err != nil {
			return fmt.Errorf("delete private original: %w", err)
		}
	}

	// A failed worker may have uploaded only a subset of variants. Listing the
	// versioned prefix instead of assuming three files makes cleanup complete
	// for both partial failures and future processing profiles.
	paginator := s3.NewListObjectsV2Paginator(deleter.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(deleter.derivativesBucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list derivatives: %w", err)
		}
		objects := make([]s3types.ObjectIdentifier, 0, len(page.Contents))
		for _, object := range page.Contents {
			objects = append(objects, s3types.ObjectIdentifier{Key: object.Key})
		}
		if len(objects) == 0 {
			continue
		}
		if _, err := deleter.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(deleter.derivativesBucket),
			Delete: &s3types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		}); err != nil {
			return fmt.Errorf("delete derivatives: %w", err)
		}
	}
	return nil
}
