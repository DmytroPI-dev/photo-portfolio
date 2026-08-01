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

func TestDynamoRepositoryCreatesDraftCollectionWithoutPublicCopies(t *testing.T) {
	collection := AdminCollection{
		Collection: Collection{ID: "sketches", Slug: "sketches", Title: "Sketches", Order: 4},
		Status:     PublicationDraft,
		Version:    1,
	}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.CreateAdminCollection(context.Background(), collection); err != nil {
		t.Fatalf("CreateAdminCollection returned error: %v", err)
	}
	if len(client.transactWriteInput.TransactItems) != 2 {
		t.Fatalf("transaction items = %d, want canonical record and admin index", len(client.transactWriteInput.TransactItems))
	}

	canonical := client.transactWriteInput.TransactItems[0].Put
	index := client.transactWriteInput.TransactItems[1].Put
	if got := attributeString(t, canonical.Item["PK"]); got != "COLLECTION#sketches" {
		t.Fatalf("canonical PK = %q, want COLLECTION#sketches", got)
	}
	if got := attributeString(t, canonical.Item["SK"]); got != canonicalAdminSortKey {
		t.Fatalf("canonical SK = %q, want ADMIN", got)
	}
	if got := attributeString(t, index.Item["PK"]); got != adminCollectionsPartition {
		t.Fatalf("index PK = %q, want %s", got, adminCollectionsPartition)
	}
}

func TestDynamoRepositoryUpdatesPublishedCollectionAndPublicCopies(t *testing.T) {
	previous := AdminCollection{
		Collection: Collection{ID: "drawings", Slug: "drawings", Title: "Drawings", Order: 1},
		Status:     PublicationPublished,
		Version:    1,
	}
	next := previous
	next.Title = "Hand Drawings"
	next.Order = 2
	next.Version = 2
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.UpdateAdminCollection(context.Background(), previous, next); err != nil {
		t.Fatalf("UpdateAdminCollection returned error: %v", err)
	}
	if len(client.transactWriteInput.TransactItems) != 6 {
		t.Fatalf("transaction items = %d, want canonical, indexes, and public copies", len(client.transactWriteInput.TransactItems))
	}

	publicMetadata := client.transactWriteInput.TransactItems[4].Put
	publicIndex := client.transactWriteInput.TransactItems[5].Put
	canonical := client.transactWriteInput.TransactItems[0].Put
	if got := attributeString(t, publicMetadata.Item["SK"]); got != metadataSortKey {
		t.Fatalf("public metadata SK = %q, want META", got)
	}
	if got := attributeString(t, publicIndex.Item["PK"]); got != collectionsPartition {
		t.Fatalf("public index PK = %q, want COLLECTIONS", got)
	}
	if got := canonical.ExpressionAttributeNames["#version"]; got != "Version" {
		t.Fatalf("version condition attribute = %q, want Version", got)
	}
}

type dynamoStub struct {
	getInput            *dynamodb.GetItemInput
	getOutput           *dynamodb.GetItemOutput
	getErr              error
	queryInput          *dynamodb.QueryInput
	queryOutput         *dynamodb.QueryOutput
	queryErr            error
	transactWriteInput  *dynamodb.TransactWriteItemsInput
	transactWriteOutput *dynamodb.TransactWriteItemsOutput
	transactWriteErr    error
}

func (stub *dynamoStub) GetItem(_ context.Context, input *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	stub.getInput = input
	return stub.getOutput, stub.getErr
}

func (stub *dynamoStub) Query(_ context.Context, input *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	stub.queryInput = input
	return stub.queryOutput, stub.queryErr
}

func (stub *dynamoStub) TransactWriteItems(_ context.Context, input *dynamodb.TransactWriteItemsInput, _ ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	stub.transactWriteInput = input
	return stub.transactWriteOutput, stub.transactWriteErr
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
