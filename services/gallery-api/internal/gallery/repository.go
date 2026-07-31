package gallery

import (
	"sort"
)

// Repository is the only data dependency used by the public HTTP layer.
// The in-memory implementation is intentionally small for Phase 1; a
// DynamoDB implementation can satisfy this same interface later.
type Repository interface {
	ListCollections() []Collection
	GetCollectionBySlug(slug string) (Collection, bool)
	ListPhotosByCollection(collectionID string) []Photo
	GetPhotoByID(id string) (Photo, bool)
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

func (repository *MemoryRepository) ListCollections() []Collection {
	return append([]Collection(nil), repository.collections...)
}

func (repository *MemoryRepository) GetCollectionBySlug(slug string) (Collection, bool) {
	collection, found := repository.collectionsBySlug[slug]
	return collection, found
}

func (repository *MemoryRepository) ListPhotosByCollection(collectionID string) []Photo {
	photos := make([]Photo, 0)
	for _, photo := range repository.photos {
		if photo.CollectionID == collectionID {
			photos = append(photos, photo)
		}
	}
	return photos
}

func (repository *MemoryRepository) GetPhotoByID(id string) (Photo, bool) {
	photo, found := repository.photosByID[id]
	return photo, found
}
