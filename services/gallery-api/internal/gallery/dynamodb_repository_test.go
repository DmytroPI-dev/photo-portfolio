package gallery

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestDynamoRepositoryListsCollectionsFromTheIndexPartition(t *testing.T) {
	collection := Collection{ID: "drawings", Slug: "drawings", Title: "Drawings", Order: 1}
	item := marshalTestItem(t, collection, collectionsPartition, "COLLECTION#0001#drawings")
	client := &dynamoStub{queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	collections, err := repository.ListCollections(context.Background())
	if err != nil {
		t.Fatalf("ListCollections returned error: %v", err)
	}
	if len(collections) != 1 || collections[0].Slug != "drawings" {
		t.Fatalf("collections = %#v, want drawings", collections)
	}
	if table := aws.ToString(client.queryInput.TableName); table != "gallery-metadata" {
		t.Fatalf("query table = %q, want gallery-metadata", table)
	}
	if got := attributeString(t, client.queryInput.ExpressionAttributeValues[":pk"]); got != collectionsPartition {
		t.Fatalf("query partition = %q, want %q", got, collectionsPartition)
	}
}

func TestDynamoRepositoryGetsPhotoByID(t *testing.T) {
	photo := Photo{ID: "drawing-01", Title: "Stillness", CollectionID: "drawings", Order: 1}
	client := &dynamoStub{getOutput: &dynamodb.GetItemOutput{Item: marshalTestItem(t, photo, photoPartition(photo.ID), metadataSortKey)}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	actual, found, err := repository.GetPhotoByID(context.Background(), photo.ID)
	if err != nil {
		t.Fatalf("GetPhotoByID returned error: %v", err)
	}
	if !found || actual.Title != "Stillness" {
		t.Fatalf("photo = %#v, found = %t, want Stillness and true", actual, found)
	}
	if got := attributeString(t, client.getInput.Key["PK"]); got != "PHOTO#drawing-01" {
		t.Fatalf("photo PK = %q, want PHOTO#drawing-01", got)
	}
}

func TestSeedWriteRequestsContainPublicAndCanonicalReadModels(t *testing.T) {
	requests, err := SeedWriteRequests()
	if err != nil {
		t.Fatalf("SeedWriteRequests returned error: %v", err)
	}

	// Each collection and photo has two public copies plus a canonical record
	// and an administrator-list index copy. A count change forces an explicit
	// review of this intentional denormalization.
	if len(requests) != 76 {
		t.Fatalf("request count = %d, want 76", len(requests))
	}

	keys := make(map[string]bool, len(requests))
	for _, request := range requests {
		item := request.PutRequest.Item
		keys[attributeString(t, item["PK"])+"|"+attributeString(t, item["SK"])] = true
	}

	for _, expected := range []string{
		"COLLECTIONS|COLLECTION#0001#drawings",
		"COLLECTION#nature|META",
		"COLLECTION#travel|PHOTO#0016#travel-05",
		"PHOTO#drawing-01|META",
		"COLLECTION#drawings|ADMIN",
		"PHOTO#drawing-01|ADMIN",
		"ADMIN#COLLECTIONS|COLLECTION#0001#drawings",
		"ADMIN#PHOTOS|PHOTO#drawings#0001#drawing-01",
	} {
		if !keys[expected] {
			t.Errorf("seed records do not include %s", expected)
		}
	}
}

func TestDynamoRepositoryReadsCanonicalAdminRecords(t *testing.T) {
	photo := AdminPhoto{
		Photo:            Photo{ID: "drawing-01", Title: "Stillness", CollectionID: "drawings", Order: 1},
		Status:           PublicationDraft,
		ProcessingStatus: ProcessingPending,
		AltText:          "A drawing of a bird",
	}
	client := &dynamoStub{getOutput: &dynamodb.GetItemOutput{Item: marshalTestItem(t, photo, photoPartition(photo.ID), canonicalAdminSortKey)}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	actual, found, err := repository.GetAdminPhotoByID(context.Background(), photo.ID)
	if err != nil {
		t.Fatalf("GetAdminPhotoByID returned error: %v", err)
	}
	if !found || actual.Status != PublicationDraft || actual.ProcessingStatus != ProcessingPending || actual.AltText != photo.AltText {
		t.Fatalf("admin photo = %#v, found = %t, want canonical private metadata", actual, found)
	}
	if got := attributeString(t, client.getInput.Key["SK"]); got != canonicalAdminSortKey {
		t.Fatalf("admin photo SK = %q, want %q", got, canonicalAdminSortKey)
	}
}

type dynamoStub struct {
	getInput    *dynamodb.GetItemInput
	getOutput   *dynamodb.GetItemOutput
	getErr      error
	queryInput  *dynamodb.QueryInput
	queryOutput *dynamodb.QueryOutput
	queryErr    error
}

func (stub *dynamoStub) GetItem(_ context.Context, input *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	stub.getInput = input
	return stub.getOutput, stub.getErr
}

func (stub *dynamoStub) Query(_ context.Context, input *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	stub.queryInput = input
	return stub.queryOutput, stub.queryErr
}

func marshalTestItem(t *testing.T, value any, partition, sort string) map[string]types.AttributeValue {
	t.Helper()
	item, err := attributevalue.MarshalMap(value)
	if err != nil {
		t.Fatalf("marshal test item: %v", err)
	}
	item["PK"] = &types.AttributeValueMemberS{Value: partition}
	item["SK"] = &types.AttributeValueMemberS{Value: sort}
	return item
}

func attributeString(t *testing.T, value types.AttributeValue) string {
	t.Helper()
	stringValue, ok := value.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatalf("attribute value = %#v, want string", value)
	}
	return stringValue.Value
}
