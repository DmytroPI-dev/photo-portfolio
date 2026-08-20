package gallery

import (
	"context"
	"errors"
	"fmt"
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

func TestDynamoRepositoryListsAdminPhotosForOneCollection(t *testing.T) {
	photo := AdminPhoto{
		Photo:  Photo{ID: "drawing-01", CollectionID: "drawings", Order: 1},
		Status: PublicationDraft,
	}
	client := &dynamoStub{queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
		marshalTestItem(t, photo, adminPhotosPartition, "PHOTO#drawings#0001#drawing-01"),
	}}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	photos, err := repository.ListAdminPhotosByCollection(context.Background(), "drawings")
	if err != nil {
		t.Fatalf("ListAdminPhotosByCollection returned error: %v", err)
	}
	if len(photos) != 1 || photos[0].ID != "drawing-01" {
		t.Fatalf("photos = %#v, want drawing-01", photos)
	}
	if got := attributeString(t, client.queryInput.ExpressionAttributeValues[":skPrefix"]); got != "PHOTO#drawings#" {
		t.Fatalf("query sort prefix = %q, want drawings prefix", got)
	}
}

func TestDynamoRepositoryPaginatesAdminPhotosForOneCollection(t *testing.T) {
	firstPhoto := AdminPhoto{Photo: Photo{ID: "drawing-01", CollectionID: "drawings", Order: 1}, Status: PublicationDraft}
	secondPhoto := AdminPhoto{Photo: Photo{ID: "drawing-02", CollectionID: "drawings", Order: 2}, Status: PublicationPublished}
	lastEvaluatedKey := key(adminPhotosPartition, "PHOTO#drawings#0001#drawing-01")
	client := &dynamoStub{queryOutputs: []*dynamodb.QueryOutput{
		{
			Items:            []map[string]types.AttributeValue{marshalTestItem(t, firstPhoto, adminPhotosPartition, "PHOTO#drawings#0001#drawing-01")},
			LastEvaluatedKey: lastEvaluatedKey,
		},
		{
			Items: []map[string]types.AttributeValue{marshalTestItem(t, secondPhoto, adminPhotosPartition, "PHOTO#drawings#0002#drawing-02")},
		},
	}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	photos, err := repository.ListAdminPhotosByCollection(context.Background(), "drawings")
	if err != nil {
		t.Fatalf("ListAdminPhotosByCollection returned error: %v", err)
	}
	if len(photos) != 2 || photos[1].ID != "drawing-02" {
		t.Fatalf("photos = %#v, want both paginated photos", photos)
	}
	if len(client.queryInputs) != 2 {
		t.Fatalf("query count = %d, want 2", len(client.queryInputs))
	}
	if got := attributeString(t, client.queryInputs[1].ExclusiveStartKey["SK"]); got != "PHOTO#drawings#0001#drawing-01" {
		t.Fatalf("second query start key = %q, want first page key", got)
	}
}

func TestDynamoRepositoryPaginatesAllAdminPhotos(t *testing.T) {
	firstPhoto := AdminPhoto{Photo: Photo{ID: "drawing-01", CollectionID: "drawings", Order: 1}, Status: PublicationDraft}
	secondPhoto := AdminPhoto{Photo: Photo{ID: "nature-01", CollectionID: "nature", Order: 1}, Status: PublicationArchived}
	lastEvaluatedKey := key(adminPhotosPartition, "PHOTO#drawings#0001#drawing-01")
	client := &dynamoStub{queryOutputs: []*dynamodb.QueryOutput{
		{Items: []map[string]types.AttributeValue{marshalTestItem(t, firstPhoto, adminPhotosPartition, "PHOTO#drawings#0001#drawing-01")}, LastEvaluatedKey: lastEvaluatedKey},
		{Items: []map[string]types.AttributeValue{marshalTestItem(t, secondPhoto, adminPhotosPartition, "PHOTO#nature#0001#nature-01")}},
	}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	photos, err := repository.ListAdminPhotos(context.Background())
	if err != nil {
		t.Fatalf("ListAdminPhotos returned error: %v", err)
	}
	if len(photos) != 2 || photos[1].ID != "nature-01" {
		t.Fatalf("photos = %#v, want both paginated photos", photos)
	}
	if len(client.queryInputs) != 2 {
		t.Fatalf("query count = %d, want 2", len(client.queryInputs))
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

func TestDynamoRepositoryCreatesDraftPhotoWithoutPublicCopies(t *testing.T) {
	photo := AdminPhoto{
		Photo:            Photo{ID: "fixture-photo", Title: "Fixture", Src: "/images/1.jpg", CollectionID: "drawings", Width: 1350, Height: 1800, Order: 7},
		Status:           PublicationDraft,
		ProcessingStatus: ProcessingNotRequired,
		Version:          1,
	}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.CreateAdminPhoto(context.Background(), photo); err != nil {
		t.Fatalf("CreateAdminPhoto returned error: %v", err)
	}
	items := client.transactWriteInput.TransactItems
	if len(items) != 2 || items[0].Put == nil || items[1].Put == nil {
		t.Fatalf("transaction items = %#v, want canonical and private index puts", items)
	}
	if got := attributeString(t, items[0].Put.Item["PK"]); got != "PHOTO#fixture-photo" {
		t.Fatalf("canonical PK = %q, want PHOTO#fixture-photo", got)
	}
	if got := attributeString(t, items[1].Put.Item["PK"]); got != adminPhotosPartition {
		t.Fatalf("index PK = %q, want %s", got, adminPhotosPartition)
	}
}

func TestDynamoRepositoryClassifiesConditionalTransactionFailuresByOperation(t *testing.T) {
	cancellation := &types.TransactionCanceledException{
		CancellationReasons: []types.CancellationReason{{Code: aws.String("ConditionalCheckFailed")}},
	}

	if err := writeError("create admin collection", cancellation, ErrAlreadyExists); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("create error = %v, want ErrAlreadyExists", err)
	}
	if err := writeError("update admin collection", cancellation, ErrVersionConflict); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("update error = %v, want ErrVersionConflict", err)
	}
}

func TestDynamoRepositoryPreservesOtherTransactionCancellationErrors(t *testing.T) {
	cancellation := &types.TransactionCanceledException{
		CancellationReasons: []types.CancellationReason{{Code: aws.String("TransactionConflict")}},
	}
	err := writeError("update admin collection", cancellation, ErrVersionConflict)
	if !errors.Is(err, cancellation) {
		t.Fatalf("error = %v, want original transaction cancellation", err)
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

func TestDynamoRepositoryReconcilesPublicCopiesForPublicationStateTransitions(t *testing.T) {
	t.Run("publishes a draft", func(t *testing.T) {
		previous := AdminCollection{
			Collection: Collection{ID: "sketches", Slug: "sketches", Title: "Sketches", Order: 4},
			Status:     PublicationDraft,
			Version:    1,
		}
		next := previous
		next.Status = PublicationPublished
		next.Version = 2

		client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
		repository := NewDynamoRepository(client, "gallery-metadata")
		if err := repository.UpdateAdminCollection(context.Background(), previous, next); err != nil {
			t.Fatalf("UpdateAdminCollection returned error: %v", err)
		}

		items := client.transactWriteInput.TransactItems
		if len(items) != 4 {
			t.Fatalf("transaction items = %d, want canonical, admin index, and two public copies", len(items))
		}
		if items[2].Put == nil || attributeString(t, items[2].Put.Item["SK"]) != metadataSortKey {
			t.Fatalf("third item = %#v, want public metadata put", items[2])
		}
		if items[3].Put == nil || attributeString(t, items[3].Put.Item["PK"]) != collectionsPartition {
			t.Fatalf("fourth item = %#v, want public index put", items[3])
		}
	})

	t.Run("archives a published collection", func(t *testing.T) {
		previous := AdminCollection{
			Collection: Collection{ID: "sketches", Slug: "sketches", Title: "Sketches", Order: 4},
			Status:     PublicationPublished,
			Version:    2,
		}
		next := previous
		next.Status = PublicationArchived
		next.Version = 3

		client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
		repository := NewDynamoRepository(client, "gallery-metadata")
		if err := repository.UpdateAdminCollection(context.Background(), previous, next); err != nil {
			t.Fatalf("UpdateAdminCollection returned error: %v", err)
		}

		items := client.transactWriteInput.TransactItems
		if len(items) != 4 {
			t.Fatalf("transaction items = %d, want canonical, admin index, and two public deletes", len(items))
		}
		if items[2].Delete == nil || attributeString(t, items[2].Delete.Key["SK"]) != metadataSortKey {
			t.Fatalf("third item = %#v, want public metadata delete", items[2])
		}
		if items[3].Delete == nil || attributeString(t, items[3].Delete.Key["PK"]) != collectionsPartition {
			t.Fatalf("fourth item = %#v, want public index delete", items[3])
		}
	})
}

func TestDynamoRepositoryPublishesPhotoAndUpdatesPublicCopies(t *testing.T) {
	previous := AdminPhoto{
		Photo:            Photo{ID: "fixture-photo", Title: "Fixture", Src: "/images/1.jpg", CollectionID: "drawings", Width: 1350, Height: 1800, Order: 7},
		Status:           PublicationDraft,
		ProcessingStatus: ProcessingNotRequired,
		Version:          1,
	}
	next := previous
	next.Status = PublicationPublished
	next.Version = 2
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.UpdateAdminPhoto(context.Background(), previous, next); err != nil {
		t.Fatalf("UpdateAdminPhoto returned error: %v", err)
	}
	items := client.transactWriteInput.TransactItems
	if len(items) != 4 {
		t.Fatalf("transaction items = %d, want canonical, admin index, and two public puts", len(items))
	}
	if got := attributeString(t, items[2].Put.Item["PK"]); got != "PHOTO#fixture-photo" {
		t.Fatalf("public metadata PK = %q, want PHOTO#fixture-photo", got)
	}
	if got := attributeString(t, items[3].Put.Item["PK"]); got != "COLLECTION#drawings" {
		t.Fatalf("public collection index PK = %q, want COLLECTION#drawings", got)
	}
	if got := items[0].Put.ExpressionAttributeNames["#version"]; got != "Version" {
		t.Fatalf("version condition attribute = %q, want Version", got)
	}
}

func TestDynamoRepositoryGuardsPublishedPhotoWritesWithCollectionVersion(t *testing.T) {
	previous := AdminPhoto{Photo: Photo{ID: "fixture-photo", Title: "Fixture", Src: "/images/1.jpg", CollectionID: "drawings", Width: 1350, Height: 1800, Order: 7}, Status: PublicationDraft, ProcessingStatus: ProcessingNotRequired, Version: 1}
	next := previous
	next.Status = PublicationPublished
	next.Version = 2
	collection := AdminCollection{Collection: Collection{ID: "drawings", Slug: "drawings", Order: 1}, Status: PublicationPublished, Version: 4}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.UpdateAdminPhotoForPublishedCollection(context.Background(), previous, next, collection); err != nil {
		t.Fatalf("UpdateAdminPhotoForPublishedCollection returned error: %v", err)
	}
	items := client.transactWriteInput.TransactItems
	guard := items[len(items)-1].ConditionCheck
	if guard == nil {
		t.Fatalf("last transaction item = %#v, want published collection condition", items[len(items)-1])
	}
	if got := attributeString(t, guard.Key["PK"]); got != "COLLECTION#drawings" {
		t.Fatalf("guard key = %q, want collection canonical key", got)
	}
	if got := attributeNumber(t, guard.ExpressionAttributeValues[":version"]); got != "4" {
		t.Fatalf("guard version = %q, want 4", got)
	}
}

func TestDynamoRepositorySplitsLargePhotoReordersIntoSafeTransactions(t *testing.T) {
	previous := make([]AdminPhoto, 18)
	next := make([]AdminPhoto, 18)
	for index := range previous {
		photo := AdminPhoto{Photo: Photo{ID: fmt.Sprintf("photo-%02d", index), Title: "Fixture", Src: "/images/1.jpg", CollectionID: "drawings", Width: 1350, Height: 1800, Order: index + 1}, Status: PublicationPublished, ProcessingStatus: ProcessingNotRequired, Version: 1}
		previous[index] = photo
		next[index] = photo
		next[index].Order = len(previous) - index
		next[index].Version = 2
	}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.ReorderAdminPhotos(context.Background(), previous, next); err != nil {
		t.Fatalf("ReorderAdminPhotos returned error: %v", err)
	}
	if len(client.transactWriteInputs) != 2 {
		t.Fatalf("transaction count = %d, want 2", len(client.transactWriteInputs))
	}
	for _, input := range client.transactWriteInputs {
		if len(input.TransactItems) > 100 {
			t.Fatalf("transaction contains %d items, want at most 100", len(input.TransactItems))
		}
	}
}

func TestDynamoRepositoryDeletesArchivedCollectionWithVersionCondition(t *testing.T) {
	collection := AdminCollection{
		Collection: Collection{ID: "sketches", Slug: "sketches", Title: "Sketches", Order: 4},
		Status:     PublicationArchived,
		Version:    3,
	}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.DeleteAdminCollection(context.Background(), collection); err != nil {
		t.Fatalf("DeleteAdminCollection returned error: %v", err)
	}
	items := client.transactWriteInput.TransactItems
	if len(items) != 2 || items[0].Delete == nil || items[1].Delete == nil {
		t.Fatalf("transaction items = %#v, want canonical and index deletes", items)
	}
	canonical := items[0].Delete
	if got := attributeString(t, canonical.Key["PK"]); got != "COLLECTION#sketches" {
		t.Fatalf("canonical delete PK = %q, want COLLECTION#sketches", got)
	}
	if got := canonical.ExpressionAttributeNames["#version"]; got != "Version" {
		t.Fatalf("version attribute = %q, want Version", got)
	}
	if got := attributeNumber(t, canonical.ExpressionAttributeValues[":version"]); got != "3" {
		t.Fatalf("version condition = %q, want 3", got)
	}
}

func TestDynamoRepositoryDeletesArchivedPhotoWithVersionCondition(t *testing.T) {
	photo := AdminPhoto{
		Photo:   Photo{ID: "photo-123", CollectionID: "drawings", Order: 7},
		Status:  PublicationArchived,
		Version: 4,
	}
	client := &dynamoStub{transactWriteOutput: &dynamodb.TransactWriteItemsOutput{}}
	repository := NewDynamoRepository(client, "gallery-metadata")

	if err := repository.DeleteAdminPhoto(context.Background(), photo); err != nil {
		t.Fatalf("DeleteAdminPhoto returned error: %v", err)
	}
	items := client.transactWriteInput.TransactItems
	if len(items) != 2 || items[0].Delete == nil || items[1].Delete == nil {
		t.Fatalf("transaction items = %#v, want canonical and index deletes", items)
	}
	canonical := items[0].Delete
	if got := attributeString(t, canonical.Key["PK"]); got != "PHOTO#photo-123" {
		t.Fatalf("canonical delete PK = %q, want PHOTO#photo-123", got)
	}
	if got := attributeNumber(t, canonical.ExpressionAttributeValues[":version"]); got != "4" {
		t.Fatalf("version condition = %q, want 4", got)
	}
	if got := attributeString(t, items[1].Delete.Key["SK"]); got != "PHOTO#drawings#0007#photo-123" {
		t.Fatalf("admin index SK = %q, want private photo index", got)
	}
}

type dynamoStub struct {
	getInput            *dynamodb.GetItemInput
	getOutput           *dynamodb.GetItemOutput
	getErr              error
	queryInput          *dynamodb.QueryInput
	queryOutput         *dynamodb.QueryOutput
	queryErr            error
	queryInputs         []*dynamodb.QueryInput
	queryOutputs        []*dynamodb.QueryOutput
	transactWriteInput  *dynamodb.TransactWriteItemsInput
	transactWriteInputs []*dynamodb.TransactWriteItemsInput
	transactWriteOutput *dynamodb.TransactWriteItemsOutput
	transactWriteErr    error
}

func (stub *dynamoStub) GetItem(_ context.Context, input *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	stub.getInput = input
	return stub.getOutput, stub.getErr
}

func (stub *dynamoStub) Query(_ context.Context, input *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	stub.queryInput = input
	stub.queryInputs = append(stub.queryInputs, input)
	if queryIndex := len(stub.queryInputs) - 1; queryIndex < len(stub.queryOutputs) {
		return stub.queryOutputs[queryIndex], nil
	}
	return stub.queryOutput, stub.queryErr
}

func (stub *dynamoStub) TransactWriteItems(_ context.Context, input *dynamodb.TransactWriteItemsInput, _ ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	stub.transactWriteInput = input
	stub.transactWriteInputs = append(stub.transactWriteInputs, input)
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

func attributeNumber(t *testing.T, value types.AttributeValue) string {
	t.Helper()
	numberValue, ok := value.(*types.AttributeValueMemberN)
	if !ok {
		t.Fatalf("attribute value = %#v, want number", value)
	}
	return numberValue.Value
}
