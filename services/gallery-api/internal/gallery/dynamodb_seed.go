package gallery

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const dynamoBatchWriteLimit = 25

// BatchWriteAPI is kept separate from DynamoDBAPI because the deployed Lambda
// never writes public metadata. Only the explicit bootstrap command receives
// credentials capable of running this operation.
type BatchWriteAPI interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// SeedDynamo writes a deterministic representation of the current placeholder
// portfolio. Put requests overwrite the same keys on every run, making the
// initial bootstrap retry-safe. Do not run it after an administrator has made
// real edits without deliberately choosing to reset that data.
func SeedDynamo(ctx context.Context, client BatchWriteAPI, table string) (int, error) {
	requests, err := SeedWriteRequests()
	if err != nil {
		return 0, err
	}

	for start := 0; start < len(requests); start += dynamoBatchWriteLimit {
		end := min(start+dynamoBatchWriteLimit, len(requests))
		if err := writeBatch(ctx, client, table, requests[start:end]); err != nil {
			return 0, err
		}
	}

	return len(requests), nil
}

// SeedWriteRequests builds four explicit access patterns: ordered collection
// list, collection metadata by slug, ordered collection photos, and a photo by
// ID. These copies remove the need for a table scan or an index at this scale.
func SeedWriteRequests() ([]types.WriteRequest, error) {
	collections, photos := SeedData()
	requests := make([]types.WriteRequest, 0, len(collections)*2+len(photos)*2)

	for _, collection := range collections {
		indexItem, err := marshalItem(collection, collectionsPartition, fmt.Sprintf("COLLECTION#%04d#%s", collection.Order, collection.Slug))
		if err != nil {
			return nil, fmt.Errorf("marshal collection index %q: %w", collection.ID, err)
		}
		metadataItem, err := marshalItem(collection, collectionPartition(collection.ID), metadataSortKey)
		if err != nil {
			return nil, fmt.Errorf("marshal collection metadata %q: %w", collection.ID, err)
		}
		requests = append(requests, putRequest(indexItem), putRequest(metadataItem))
	}

	for _, photo := range photos {
		collectionItem, err := marshalItem(photo, collectionPartition(photo.CollectionID), fmt.Sprintf("PHOTO#%04d#%s", photo.Order, photo.ID))
		if err != nil {
			return nil, fmt.Errorf("marshal collection photo %q: %w", photo.ID, err)
		}
		metadataItem, err := marshalItem(photo, photoPartition(photo.ID), metadataSortKey)
		if err != nil {
			return nil, fmt.Errorf("marshal photo metadata %q: %w", photo.ID, err)
		}
		requests = append(requests, putRequest(collectionItem), putRequest(metadataItem))
	}

	return requests, nil
}

func marshalItem(value any, partition, sort string) (map[string]types.AttributeValue, error) {
	item, err := attributevalue.MarshalMap(value)
	if err != nil {
		return nil, err
	}
	item["PK"] = &types.AttributeValueMemberS{Value: partition}
	item["SK"] = &types.AttributeValueMemberS{Value: sort}
	return item, nil
}

func putRequest(item map[string]types.AttributeValue) types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: item}}
}

func writeBatch(ctx context.Context, client BatchWriteAPI, table string, requests []types.WriteRequest) error {
	pending := requests
	for attempt := range 5 {
		output, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{table: pending},
		})
		if err != nil {
			return fmt.Errorf("write metadata batch: %w", err)
		}

		pending = output.UnprocessedItems[table]
		if len(pending) == 0 {
			return nil
		}

		// DynamoDB may throttle a batch even when on-demand capacity is used.
		// Brief incremental backoff keeps this tiny bootstrap reliable without
		// adding a retry framework or hiding a persistent failure indefinitely.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
		}
	}

	return fmt.Errorf("write metadata batch: %d requests remained after retries", len(pending))
}
