// Package storage owns AWS integrations that move image work between the API
// and the worker. The API enqueues the same S3-shaped message emitted for a
// normal upload, so the worker has exactly one input contract to maintain.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ProcessingQueue accepts a request to process one private original. Keeping
// this interface small lets the HTTP package test retry behavior without an
// AWS client or a queue emulator.
type ProcessingQueue interface {
	EnqueueOriginal(ctx context.Context, originalKey string) error
}

type sqsMessageSender interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// SQSProcessingQueue creates a worker-compatible S3 notification and sends it
// to the durable image-processing queue. It deliberately does not update photo
// metadata: the worker remains the only component that changes processing
// status after it has successfully started or finished a transformation.
type SQSProcessingQueue struct {
	client          sqsMessageSender
	queueURL        string
	originalsBucket string
}

func NewSQSProcessingQueue(awsConfig aws.Config, queueURL, originalsBucket string) *SQSProcessingQueue {
	return &SQSProcessingQueue{
		client:          sqs.NewFromConfig(awsConfig),
		queueURL:        queueURL,
		originalsBucket: originalsBucket,
	}
}

func (queue *SQSProcessingQueue) EnqueueOriginal(ctx context.Context, originalKey string) error {
	message, err := s3ObjectCreatedMessage(queue.originalsBucket, originalKey)
	if err != nil {
		return err
	}
	_, err = queue.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue.queueURL),
		MessageBody: aws.String(string(message)),
	})
	if err != nil {
		return fmt.Errorf("send image-processing message: %w", err)
	}
	return nil
}

// s3ObjectCreatedMessage reproduces the fields consumed by processing.Worker.
// S3 URL-encodes object keys in event payloads, including slashes, and the
// worker decodes them before looking up the draft's OriginalKey.
func s3ObjectCreatedMessage(bucket, key string) ([]byte, error) {
	type object struct {
		Key string `json:"key"`
	}
	type bucketValue struct {
		Name string `json:"name"`
	}
	type s3 struct {
		Bucket bucketValue `json:"bucket"`
		Object object      `json:"object"`
	}
	type record struct {
		EventSource string `json:"eventSource"`
		EventName   string `json:"eventName"`
		S3          s3     `json:"s3"`
	}

	return json.Marshal(struct {
		Records []record `json:"Records"`
	}{Records: []record{{
		EventSource: "aws:s3",
		EventName:   "ObjectCreated:Put",
		S3: s3{
			Bucket: bucketValue{Name: bucket},
			Object: object{Key: url.QueryEscape(key)},
		},
	}}})
}
