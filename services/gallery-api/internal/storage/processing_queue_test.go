package storage

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestS3ObjectCreatedMessageIsWorkerCompatible(t *testing.T) {
	message, err := s3ObjectCreatedMessage("originals.example.test", "originals/photo-123/original image.jpg")
	if err != nil {
		t.Fatalf("create S3 event message: %v", err)
	}

	var event events.S3Event
	if err := json.Unmarshal(message, &event); err != nil {
		t.Fatalf("unmarshal worker event: %v", err)
	}
	if len(event.Records) != 1 || event.Records[0].S3.Bucket.Name != "originals.example.test" {
		t.Fatalf("event records = %#v, want one original-bucket event", event.Records)
	}
	key, err := url.QueryUnescape(event.Records[0].S3.Object.Key)
	if err != nil {
		t.Fatalf("decode object key: %v", err)
	}
	if key != "originals/photo-123/original image.jpg" {
		t.Fatalf("decoded key = %q, want original key", key)
	}
}
