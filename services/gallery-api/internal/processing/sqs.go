package processing

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
)

// HandleSQSEvent unwraps S3's notification envelope from SQS messages. The
// mapping starts with batch_size = 1, but partial batch failures are still
// reported so a later throughput adjustment never replays successful images.
func (worker *Worker) HandleSQSEvent(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	response := events.SQSEventResponse{}
	for _, message := range event.Records {
		var notification events.S3Event
		if err := json.Unmarshal([]byte(message.Body), &notification); err != nil {
			response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: message.MessageId})
			continue
		}
		messageFailed := false
		for _, record := range notification.Records {
			if record.S3.Bucket.Name != worker.originalsBucket {
				messageFailed = true
				continue
			}
			objectKey, err := url.QueryUnescape(record.S3.Object.Key)
			if err != nil {
				messageFailed = true
				continue
			}
			if err := worker.ProcessObject(ctx, objectKey); err != nil {
				messageFailed = true
			}
		}
		if messageFailed {
			response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: message.MessageId})
		}
	}
	return response, nil
}
