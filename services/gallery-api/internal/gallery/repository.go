package gallery

import (
	"context"
	"sort"
)

// Repository is the only data dependency used by the public HTTP layer.
// The in-memory implementation is intentionally small for Phase 1; a
// DynamoDB implementation can satisfy this same interface later.
type Repository interface {
	ListCollections(ctx context.Context) ([]Collection, error)
	GetCollectionBySlug(ctx context.Context, slug string) (Collection, bool, error)
	ListPhotosByCollection(ctx context.Context, collectionID string) ([]Photo, error)
	GetPhotoByID(ctx context.Context, id string) (Photo, bool, error)
}

type MemoryRepository struct {
	collections       []Collection
	collectionsBySlug map[string]Collection
	photos            []Photo
	photosByID        map[string]Photo
}

func NewMemoryRepository(collections []Collection, photos []Photo) *MemoryRepository {
	repository := &MemoryRepository{
		collections:       append([]Collection(nil), collections...),
		collectionsBySlug: make(map[string]Collection, len(collections)),
		photos:            append([]Photo(nil), photos...),
		photosByID:        make(map[string]Photo, len(photos)),
	}

	// The API owns ordering rather than relying on a caller's insertion order.
	// DynamoDB will eventually reproduce this through the collection sort key.
	sort.Slice(repository.collections, func(i, j int) bool {
		return repository.collections[i].Order < repository.collections[j].Order
	})
	sort.Slice(repository.photos, func(i, j int) bool {
		return repository.photos[i].Order < repository.photos[j].Order
	})

	for _, collection := range repository.collections {
		repository.collectionsBySlug[collection.Slug] = collection
	}
	for _, photo := range repository.photos {
		repository.photosByID[photo.ID] = photo
	}

	return repository
}

func (repository *MemoryRepository) ListCollections(_ context.Context) ([]Collection, error) {
	return append([]Collection(nil), repository.collections...), nil
}

func (repository *MemoryRepository) GetCollectionBySlug(_ context.Context, slug string) (Collection, bool, error) {
	collection, found := repository.collectionsBySlug[slug]
	return collection, found, nil
}

func (repository *MemoryRepository) ListPhotosByCollection(_ context.Context, collectionID string) ([]Photo, error) {
	photos := make([]Photo, 0)
	for _, photo := range repository.photos {
		if photo.CollectionID == collectionID {
			photos = append(photos, photo)
		}
	}
	return photos, nil
}

func (repository *MemoryRepository) GetPhotoByID(_ context.Context, id string) (Photo, bool, error) {
	photo, found := repository.photosByID[id]
	return photo, found, nil
}
