package gallery

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	collectionsPartition      = "COLLECTIONS"
	metadataSortKey           = "META"
	canonicalAdminSortKey     = "ADMIN"
	adminCollectionsPartition = "ADMIN#COLLECTIONS"
	adminPhotosPartition      = "ADMIN#PHOTOS"
)

// DynamoDBAPI describes the small part of the AWS client required by public
// gallery reads. Tests use a stub of this interface, so key design and mapping
// can be verified without credentials or a DynamoDB Local process.
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

// DynamoRepository reads the deliberately duplicated DynamoDB records
// documented in the implementation plan. Each public endpoint therefore needs
// only one GetItem or Query operation, with no Scan or secondary index.
type DynamoRepository struct {
	client DynamoDBAPI
	table  string
}

func NewDynamoRepository(client DynamoDBAPI, table string) *DynamoRepository {
	return &DynamoRepository{client: client, table: table}
}

func (repository *DynamoRepository) ListCollections(ctx context.Context) ([]Collection, error) {
	output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repository.table),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: collectionsPartition},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query collection index: %w", err)
	}

	collections := make([]Collection, 0, len(output.Items))
	for _, item := range output.Items {
		var collection Collection
		if err := attributevalue.UnmarshalMap(item, &collection); err != nil {
			return nil, fmt.Errorf("decode collection index item: %w", err)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

func (repository *DynamoRepository) GetCollectionBySlug(ctx context.Context, slug string) (Collection, bool, error) {
	output, err := repository.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repository.table),
		Key:       key(collectionPartition(slug), metadataSortKey),
	})
	if err != nil {
		return Collection{}, false, fmt.Errorf("get collection %q: %w", slug, err)
	}
	if len(output.Item) == 0 {
		return Collection{}, false, nil
	}

	var collection Collection
	if err := attributevalue.UnmarshalMap(output.Item, &collection); err != nil {
		return Collection{}, false, fmt.Errorf("decode collection %q: %w", slug, err)
	}
	return collection, true, nil
}

func (repository *DynamoRepository) ListPhotosByCollection(ctx context.Context, collectionID string) ([]Photo, error) {
	output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repository.table),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: collectionPartition(collectionID)},
			":skPrefix": &types.AttributeValueMemberS{Value: "PHOTO#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query photos for collection %q: %w", collectionID, err)
	}

	photos := make([]Photo, 0, len(output.Items))
	for _, item := range output.Items {
		var photo Photo
		if err := attributevalue.UnmarshalMap(item, &photo); err != nil {
			return nil, fmt.Errorf("decode collection photo: %w", err)
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

func (repository *DynamoRepository) GetPhotoByID(ctx context.Context, id string) (Photo, bool, error) {
	output, err := repository.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repository.table),
		Key:       key(photoPartition(id), metadataSortKey),
	})
	if err != nil {
		return Photo{}, false, fmt.Errorf("get photo %q: %w", id, err)
	}
	if len(output.Item) == 0 {
		return Photo{}, false, nil
	}

	var photo Photo
	if err := attributevalue.UnmarshalMap(output.Item, &photo); err != nil {
		return Photo{}, false, fmt.Errorf("decode photo %q: %w", id, err)
	}
	return photo, true, nil
}

// ListAdminCollections reads the private ordered index, rather than deriving
// the result from public collection copies. The method is not exposed by an
// HTTP route yet; it is the repository foundation for CRUD forms and drafts.
func (repository *DynamoRepository) ListAdminCollections(ctx context.Context) ([]AdminCollection, error) {
	output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repository.table),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: adminCollectionsPartition},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query admin collection index: %w", err)
	}

	collections := make([]AdminCollection, 0, len(output.Items))
	for _, item := range output.Items {
		var collection AdminCollection
		if err := attributevalue.UnmarshalMap(item, &collection); err != nil {
			return nil, fmt.Errorf("decode admin collection index item: %w", err)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

// GetAdminCollectionByID resolves the canonical collection record. It uses a
// different sort key from the public META record under the same partition.
func (repository *DynamoRepository) GetAdminCollectionByID(ctx context.Context, id string) (AdminCollection, bool, error) {
	output, err := repository.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repository.table),
		Key:       key(collectionPartition(id), canonicalAdminSortKey),
	})
	if err != nil {
		return AdminCollection{}, false, fmt.Errorf("get admin collection %q: %w", id, err)
	}
	if len(output.Item) == 0 {
		return AdminCollection{}, false, nil
	}

	var collection AdminCollection
	if err := attributevalue.UnmarshalMap(output.Item, &collection); err != nil {
		return AdminCollection{}, false, fmt.Errorf("decode admin collection %q: %w", id, err)
	}
	return collection, true, nil
}

// ListAdminPhotos reads one dedicated index for all administrator-visible
// photos, including drafts and archived work. Its sort key groups photos by
// collection and preserves each collection's display order.
func (repository *DynamoRepository) ListAdminPhotos(ctx context.Context) ([]AdminPhoto, error) {
	output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repository.table),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: adminPhotosPartition},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query admin photo index: %w", err)
	}

	photos := make([]AdminPhoto, 0, len(output.Items))
	for _, item := range output.Items {
		var photo AdminPhoto
		if err := attributevalue.UnmarshalMap(item, &photo); err != nil {
			return nil, fmt.Errorf("decode admin photo index item: %w", err)
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

// GetAdminPhotoByID resolves the canonical photo record, including metadata
// that is intentionally absent from the public gallery response.
func (repository *DynamoRepository) GetAdminPhotoByID(ctx context.Context, id string) (AdminPhoto, bool, error) {
	output, err := repository.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repository.table),
		Key:       key(photoPartition(id), canonicalAdminSortKey),
	})
	if err != nil {
		return AdminPhoto{}, false, fmt.Errorf("get admin photo %q: %w", id, err)
	}
	if len(output.Item) == 0 {
		return AdminPhoto{}, false, nil
	}

	var photo AdminPhoto
	if err := attributevalue.UnmarshalMap(output.Item, &photo); err != nil {
		return AdminPhoto{}, false, fmt.Errorf("decode admin photo %q: %w", id, err)
	}
	return photo, true, nil
}

// CreateAdminCollection writes only canonical data. New collections begin as
// drafts, so no anonymous public read records exist until a later publish
// action explicitly creates them.
func (repository *DynamoRepository) CreateAdminCollection(ctx context.Context, collection AdminCollection) error {
	canonicalItem, err := marshalItem(collection, collectionPartition(collection.ID), canonicalAdminSortKey)
	if err != nil {
		return fmt.Errorf("marshal canonical collection %q: %w", collection.ID, err)
	}
	indexItem, err := marshalItem(collection, adminCollectionsPartition, adminCollectionIndexSortKey(collection))
	if err != nil {
		return fmt.Errorf("marshal admin collection index %q: %w", collection.ID, err)
	}

	_, err = repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			conditionalPut(repository.table, canonicalItem, "attribute_not_exists(PK)"),
			conditionalPut(repository.table, indexItem, "attribute_not_exists(PK)"),
		},
	})
	return writeError("create admin collection", err)
}

// UpdateAdminCollection replaces the canonical document and its ordered index
// atomically. A version condition prevents two browser tabs from silently
// overwriting each other's metadata. Published collections update their two
// public copies in the same transaction; drafts have no public copies yet.
func (repository *DynamoRepository) UpdateAdminCollection(ctx context.Context, previous, next AdminCollection) error {
	canonicalItem, err := marshalItem(next, collectionPartition(next.ID), canonicalAdminSortKey)
	if err != nil {
		return fmt.Errorf("marshal canonical collection %q: %w", next.ID, err)
	}
	newIndexItem, err := marshalItem(next, adminCollectionsPartition, adminCollectionIndexSortKey(next))
	if err != nil {
		return fmt.Errorf("marshal admin collection index %q: %w", next.ID, err)
	}

	items := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName:           aws.String(repository.table),
				Item:                canonicalItem,
				ConditionExpression: aws.String("#version = :version"),
				ExpressionAttributeNames: map[string]string{
					// attributevalue uses the Go field name for DynamoDB attributes
					// unless a dynamodbav tag is present. Existing canonical seed
					// data therefore stores this as Version, not JSON's version.
					"#version": "Version",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", previous.Version)},
				},
			},
		},
	}

	oldIndexKey := adminCollectionIndexSortKey(previous)
	newIndexKey := adminCollectionIndexSortKey(next)
	if oldIndexKey != newIndexKey {
		items = append(items, deleteItem(repository.table, key(adminCollectionsPartition, oldIndexKey)))
	}
	items = append(items, putItem(repository.table, newIndexItem))

	if previous.Status == PublicationPublished {
		publicMetadata, err := marshalItem(next.Collection, collectionPartition(next.ID), metadataSortKey)
		if err != nil {
			return fmt.Errorf("marshal public collection %q: %w", next.ID, err)
		}
		publicIndex, err := marshalItem(next.Collection, collectionsPartition, publicCollectionIndexSortKey(next.Collection))
		if err != nil {
			return fmt.Errorf("marshal public collection index %q: %w", next.ID, err)
		}

		oldPublicIndexKey := publicCollectionIndexSortKey(previous.Collection)
		newPublicIndexKey := publicCollectionIndexSortKey(next.Collection)
		if oldPublicIndexKey != newPublicIndexKey {
			items = append(items, deleteItem(repository.table, key(collectionsPartition, oldPublicIndexKey)))
		}
		items = append(items, putItem(repository.table, publicMetadata), putItem(repository.table, publicIndex))
	}

	_, err = repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	return writeError("update admin collection", err)
}

func adminCollectionIndexSortKey(collection AdminCollection) string {
	return fmt.Sprintf("COLLECTION#%04d#%s", collection.Order, collection.ID)
}

func publicCollectionIndexSortKey(collection Collection) string {
	return fmt.Sprintf("COLLECTION#%04d#%s", collection.Order, collection.Slug)
}

func conditionalPut(table string, item map[string]types.AttributeValue, condition string) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(table),
		Item:                item,
		ConditionExpression: aws.String(condition),
	}}
}

func putItem(table string, item map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{TableName: aws.String(table), Item: item}}
}

func deleteItem(table string, itemKey map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Delete: &types.Delete{TableName: aws.String(table), Key: itemKey}}
}

func writeError(operation string, err error) error {
	if err == nil {
		return nil
	}

	var cancelled *types.TransactionCanceledException
	if errors.As(err, &cancelled) {
		return ErrVersionConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func key(partition, sort string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: partition},
		"SK": &types.AttributeValueMemberS{Value: sort},
	}
}

func collectionPartition(collectionID string) string {
	return "COLLECTION#" + collectionID
}

func photoPartition(photoID string) string {
	return "PHOTO#" + photoID
}
