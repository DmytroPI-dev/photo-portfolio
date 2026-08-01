package gallery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	collectionsPartition = "COLLECTIONS"
	metadataSortKey      = "META"
)

// DynamoDBAPI describes the small part of the AWS client required by public
// gallery reads. Tests use a stub of this interface, so key design and mapping
// can be verified without credentials or a DynamoDB Local process.
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
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
