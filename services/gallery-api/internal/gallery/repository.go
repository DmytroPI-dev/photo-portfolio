package gallery

import (
	"context"
	"errors"
	"sort"
)

var (
	// ErrAlreadyExists and ErrVersionConflict become stable HTTP 409 responses.
	// They keep DynamoDB's implementation-specific transaction details out of
	// both the API contract and React-admin error notifications.
	ErrAlreadyExists   = errors.New("gallery record already exists")
	ErrVersionConflict = errors.New("gallery record version conflict")
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

// AdminCollectionRepository is deliberately separate from the anonymous
// Repository contract. It reads canonical documents, including draft state and
// version metadata that must never enter public gallery responses.
type AdminCollectionRepository interface {
	ListAdminCollections(ctx context.Context) ([]AdminCollection, error)
	GetAdminCollectionByID(ctx context.Context, id string) (AdminCollection, bool, error)
	CreateAdminCollection(ctx context.Context, collection AdminCollection) error
	UpdateAdminCollection(ctx context.Context, previous, next AdminCollection) error
}

type MemoryRepository struct {
	collections       []Collection
	collectionsBySlug map[string]Collection
	photos            []Photo
	photosByID        map[string]Photo
	adminCollections  []AdminCollection
	adminByID         map[string]AdminCollection
}

func NewMemoryRepository(collections []Collection, photos []Photo) *MemoryRepository {
	repository := &MemoryRepository{
		collections:       append([]Collection(nil), collections...),
		collectionsBySlug: make(map[string]Collection, len(collections)),
		photos:            append([]Photo(nil), photos...),
		photosByID:        make(map[string]Photo, len(photos)),
		adminCollections:  make([]AdminCollection, 0, len(collections)),
		adminByID:         make(map[string]AdminCollection, len(collections)),
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
		adminCollection := AdminCollection{Collection: collection, Status: PublicationPublished, Version: 1}
		repository.adminCollections = append(repository.adminCollections, adminCollection)
		repository.adminByID[collection.ID] = adminCollection
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

func (repository *MemoryRepository) ListAdminCollections(_ context.Context) ([]AdminCollection, error) {
	collections := append([]AdminCollection(nil), repository.adminCollections...)
	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Order < collections[j].Order
	})
	return collections, nil
}

func (repository *MemoryRepository) GetAdminCollectionByID(_ context.Context, id string) (AdminCollection, bool, error) {
	collection, found := repository.adminByID[id]
	return collection, found, nil
}

func (repository *MemoryRepository) CreateAdminCollection(_ context.Context, collection AdminCollection) error {
	if _, found := repository.adminByID[collection.ID]; found {
		return ErrAlreadyExists
	}
	repository.adminCollections = append(repository.adminCollections, collection)
	repository.adminByID[collection.ID] = collection
	return nil
}

func (repository *MemoryRepository) UpdateAdminCollection(_ context.Context, previous, next AdminCollection) error {
	current, found := repository.adminByID[previous.ID]
	if !found || current.Version != previous.Version {
		return ErrVersionConflict
	}

	for index, collection := range repository.adminCollections {
		if collection.ID == previous.ID {
			repository.adminCollections[index] = next
			break
		}
	}
	repository.adminByID[next.ID] = next
	if current.Status == PublicationPublished {
		for index, collection := range repository.collections {
			if collection.ID == previous.ID {
				repository.collections[index] = next.Collection
				break
			}
		}
		repository.collectionsBySlug[next.Slug] = next.Collection
	}
	return nil
}
