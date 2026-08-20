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
	photos := make([]AdminPhoto, 0)
	var startKey map[string]types.AttributeValue
	for {
		output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(repository.table),
			KeyConditionExpression: aws.String("PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: adminPhotosPartition},
			},
			ExclusiveStartKey: startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("query admin photo index: %w", err)
		}

		for _, item := range output.Items {
			var photo AdminPhoto
			if err := attributevalue.UnmarshalMap(item, &photo); err != nil {
				return nil, fmt.Errorf("decode admin photo index item: %w", err)
			}
			photos = append(photos, photo)
		}
		if len(output.LastEvaluatedKey) == 0 {
			return photos, nil
		}
		startKey = output.LastEvaluatedKey
	}
}

// ListAdminPhotosByCollection uses the private ordered index to protect
// collection lifecycle operations from leaving draft or archived photos behind.
func (repository *DynamoRepository) ListAdminPhotosByCollection(ctx context.Context, collectionID string) ([]AdminPhoto, error) {
	photos := make([]AdminPhoto, 0)
	var startKey map[string]types.AttributeValue
	for {
		output, err := repository.client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(repository.table),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: adminPhotosPartition},
				":skPrefix": &types.AttributeValueMemberS{Value: "PHOTO#" + collectionID + "#"},
			},
			ExclusiveStartKey: startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("query admin photos for collection %q: %w", collectionID, err)
		}

		for _, item := range output.Items {
			var photo AdminPhoto
			if err := attributevalue.UnmarshalMap(item, &photo); err != nil {
				return nil, fmt.Errorf("decode admin collection photo: %w", err)
			}
			photos = append(photos, photo)
		}

		if len(output.LastEvaluatedKey) == 0 {
			return photos, nil
		}
		startKey = output.LastEvaluatedKey
	}
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
	return writeError("create admin collection", err, ErrAlreadyExists)
}

// UpdateAdminCollection replaces the canonical document and its ordered index
// atomically. A version condition prevents two browser tabs from silently
// overwriting each other's metadata. The same transaction reconciles public
// copies whenever a collection enters or leaves the published state.
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

	wasPublished := previous.Status == PublicationPublished
	isPublished := next.Status == PublicationPublished
	switch {
	case wasPublished && !isPublished:
		items = append(items,
			deleteItem(repository.table, key(collectionPartition(previous.ID), metadataSortKey)),
			deleteItem(repository.table, key(collectionsPartition, publicCollectionIndexSortKey(previous.Collection))),
		)
	case !wasPublished && isPublished:
		publicMetadata, err := marshalItem(next.Collection, collectionPartition(next.ID), metadataSortKey)
		if err != nil {
			return fmt.Errorf("marshal public collection %q: %w", next.ID, err)
		}
		publicIndex, err := marshalItem(next.Collection, collectionsPartition, publicCollectionIndexSortKey(next.Collection))
		if err != nil {
			return fmt.Errorf("marshal public collection index %q: %w", next.ID, err)
		}
		items = append(items, putItem(repository.table, publicMetadata), putItem(repository.table, publicIndex))
	case wasPublished && isPublished:
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
	return writeError("update admin collection", err, ErrVersionConflict)
}

// DeleteAdminCollection removes both private copies atomically. The handler
// guarantees the collection is archived and empty before calling this method;
// the Version condition still prevents a stale confirmation from deleting a
// record that another administrator changed in the meantime.
func (repository *DynamoRepository) DeleteAdminCollection(ctx context.Context, collection AdminCollection) error {
	_, err := repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			conditionalDelete(repository.table, key(collectionPartition(collection.ID), canonicalAdminSortKey), "#version = :version", map[string]string{"#version": "Version"}, map[string]types.AttributeValue{
				":version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", collection.Version)},
			}),
			deleteItem(repository.table, key(adminCollectionsPartition, adminCollectionIndexSortKey(collection))),
		},
	})
	return writeError("delete admin collection", err, ErrVersionConflict)
}

// CreateAdminPhoto writes the canonical record and its private list entry. A
// new photo is always a draft, so there are no anonymous gallery copies yet.
func (repository *DynamoRepository) CreateAdminPhoto(ctx context.Context, photo AdminPhoto) error {
	canonicalItem, err := marshalItem(photo, photoPartition(photo.ID), canonicalAdminSortKey)
	if err != nil {
		return fmt.Errorf("marshal canonical photo %q: %w", photo.ID, err)
	}
	indexItem, err := marshalItem(photo, adminPhotosPartition, adminPhotoIndexSortKey(photo))
	if err != nil {
		return fmt.Errorf("marshal admin photo index %q: %w", photo.ID, err)
	}

	_, err = repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			conditionalPut(repository.table, canonicalItem, "attribute_not_exists(PK)"),
			conditionalPut(repository.table, indexItem, "attribute_not_exists(PK)"),
		},
	})
	return writeError("create admin photo", err, ErrAlreadyExists)
}

// UpdateAdminPhoto keeps the canonical record, private index, and (when
// published) public metadata/index copies in one DynamoDB transaction.
func (repository *DynamoRepository) UpdateAdminPhoto(ctx context.Context, previous, next AdminPhoto) error {
	items, err := repository.adminPhotoUpdateItems(previous, next)
	if err != nil {
		return err
	}
	_, err = repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	return writeError("update admin photo", err, ErrVersionConflict)
}

// DeleteAdminPhoto removes the private canonical document and its admin-list
// entry atomically. The handler permits this only after archive has removed
// public copies and after object cleanup has completed.
func (repository *DynamoRepository) DeleteAdminPhoto(ctx context.Context, photo AdminPhoto) error {
	_, err := repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			conditionalDelete(repository.table, key(photoPartition(photo.ID), canonicalAdminSortKey), "#version = :version", map[string]string{"#version": "Version"}, map[string]types.AttributeValue{
				":version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", photo.Version)},
			}),
			deleteItem(repository.table, key(adminPhotosPartition, adminPhotoIndexSortKey(photo))),
		},
	})
	return writeError("delete admin photo", err, ErrVersionConflict)
}

func (repository *DynamoRepository) UpdateAdminPhotoForPublishedCollection(ctx context.Context, previous, next AdminPhoto, collection AdminCollection) error {
	items, err := repository.adminPhotoUpdateItems(previous, next)
	if err != nil {
		return err
	}
	items = append(items, conditionalPublishedCollection(repository.table, collection))
	_, err = repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	return writeError("update admin photo", err, ErrVersionConflict)
}

// ReorderAdminPhotos updates photos in transactions no larger than DynamoDB's
// 100-item limit. Each photo remains conditionally versioned, so callers can
// safely retry after a concurrent edit instead of being limited by collection
// size.
func (repository *DynamoRepository) ReorderAdminPhotos(ctx context.Context, previous, next []AdminPhoto) error {
	if len(previous) != len(next) {
		return ErrVersionConflict
	}
	items := make([]types.TransactWriteItem, 0, 100)
	for index, photo := range previous {
		if photo.ID != next[index].ID {
			return ErrVersionConflict
		}
		photoItems, err := repository.adminPhotoUpdateItems(photo, next[index])
		if err != nil {
			return err
		}
		if len(items)+len(photoItems) > 100 {
			if _, err := repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items}); err != nil {
				return writeError("reorder admin photos", err, ErrVersionConflict)
			}
			items = make([]types.TransactWriteItem, 0, 100)
		}
		items = append(items, photoItems...)
	}
	if len(items) == 0 {
		return nil
	}
	_, err := repository.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	return writeError("reorder admin photos", err, ErrVersionConflict)
}

func (repository *DynamoRepository) adminPhotoUpdateItems(previous, next AdminPhoto) ([]types.TransactWriteItem, error) {
	canonicalItem, err := marshalItem(next, photoPartition(next.ID), canonicalAdminSortKey)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical photo %q: %w", next.ID, err)
	}
	newIndexItem, err := marshalItem(next, adminPhotosPartition, adminPhotoIndexSortKey(next))
	if err != nil {
		return nil, fmt.Errorf("marshal admin photo index %q: %w", next.ID, err)
	}

	items := []types.TransactWriteItem{
		conditionalPutWithVersion(repository.table, canonicalItem, previous.Version),
	}
	oldIndexKey := adminPhotoIndexSortKey(previous)
	newIndexKey := adminPhotoIndexSortKey(next)
	if oldIndexKey != newIndexKey {
		items = append(items, deleteItem(repository.table, key(adminPhotosPartition, oldIndexKey)))
	}
	items = append(items, putItem(repository.table, newIndexItem))

	wasPublished := previous.Status == PublicationPublished
	isPublished := next.Status == PublicationPublished
	switch {
	case wasPublished && !isPublished:
		items = append(items,
			deleteItem(repository.table, key(photoPartition(previous.ID), metadataSortKey)),
			deleteItem(repository.table, key(collectionPartition(previous.CollectionID), publicPhotoIndexSortKey(previous.Photo))),
		)
	case !wasPublished && isPublished:
		publicItems, err := publicPhotoPutItems(repository.table, next.Photo)
		if err != nil {
			return nil, err
		}
		items = append(items, publicItems...)
	case wasPublished && isPublished:
		oldPublicKey := publicPhotoIndexSortKey(previous.Photo)
		newPublicKey := publicPhotoIndexSortKey(next.Photo)
		if previous.CollectionID != next.CollectionID || oldPublicKey != newPublicKey {
			items = append(items, deleteItem(repository.table, key(collectionPartition(previous.CollectionID), oldPublicKey)))
		}
		publicItems, err := publicPhotoPutItems(repository.table, next.Photo)
		if err != nil {
			return nil, err
		}
		items = append(items, publicItems...)
	}
	return items, nil
}

func publicPhotoPutItems(table string, photo Photo) ([]types.TransactWriteItem, error) {
	metadataItem, err := marshalItem(photo, photoPartition(photo.ID), metadataSortKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public photo %q: %w", photo.ID, err)
	}
	collectionItem, err := marshalItem(photo, collectionPartition(photo.CollectionID), publicPhotoIndexSortKey(photo))
	if err != nil {
		return nil, fmt.Errorf("marshal public collection photo %q: %w", photo.ID, err)
	}
	return []types.TransactWriteItem{putItem(table, metadataItem), putItem(table, collectionItem)}, nil
}

func adminCollectionIndexSortKey(collection AdminCollection) string {
	return fmt.Sprintf("COLLECTION#%04d#%s", collection.Order, collection.ID)
}

func publicCollectionIndexSortKey(collection Collection) string {
	return fmt.Sprintf("COLLECTION#%04d#%s", collection.Order, collection.Slug)
}

func adminPhotoIndexSortKey(photo AdminPhoto) string {
	return fmt.Sprintf("PHOTO#%s#%04d#%s", photo.CollectionID, photo.Order, photo.ID)
}

func publicPhotoIndexSortKey(photo Photo) string {
	return fmt.Sprintf("PHOTO#%04d#%s", photo.Order, photo.ID)
}

func conditionalPut(table string, item map[string]types.AttributeValue, condition string) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(table),
		Item:                item,
		ConditionExpression: aws.String(condition),
	}}
}

func conditionalPutWithVersion(table string, item map[string]types.AttributeValue, version int) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(table),
		Item:                item,
		ConditionExpression: aws.String("#version = :version"),
		ExpressionAttributeNames: map[string]string{
			"#version": "Version",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", version)},
		},
	}}
}

func conditionalPublishedCollection(table string, collection AdminCollection) types.TransactWriteItem {
	return types.TransactWriteItem{ConditionCheck: &types.ConditionCheck{
		TableName:           aws.String(table),
		Key:                 key(collectionPartition(collection.ID), canonicalAdminSortKey),
		ConditionExpression: aws.String("#status = :published AND #version = :version"),
		ExpressionAttributeNames: map[string]string{
			"#status":  "Status",
			"#version": "Version",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":published": &types.AttributeValueMemberS{Value: string(PublicationPublished)},
			":version":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", collection.Version)},
		},
	}}
}

func putItem(table string, item map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{TableName: aws.String(table), Item: item}}
}

func deleteItem(table string, itemKey map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Delete: &types.Delete{TableName: aws.String(table), Key: itemKey}}
}

func conditionalDelete(table string, itemKey map[string]types.AttributeValue, condition string, names map[string]string, values map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Delete: &types.Delete{
		TableName:                 aws.String(table),
		Key:                       itemKey,
		ConditionExpression:       aws.String(condition),
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
	}}
}

func writeError(operation string, err, conditionFailure error) error {
	if err == nil {
		return nil
	}

	var cancelled *types.TransactionCanceledException
	if errors.As(err, &cancelled) && hasConditionalCheckFailure(cancelled) {
		return conditionFailure
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func hasConditionalCheckFailure(cancelled *types.TransactionCanceledException) bool {
	for _, reason := range cancelled.CancellationReasons {
		if aws.ToString(reason.Code) == "ConditionalCheckFailed" {
			return true
		}
	}
	return false
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
