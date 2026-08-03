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
	// ListAdminPhotosByCollection checks canonical photo ownership. Lifecycle
	// actions use it instead of public copies so drafts cannot be orphaned.
	ListAdminPhotosByCollection(ctx context.Context, collectionID string) ([]AdminPhoto, error)
	CreateAdminCollection(ctx context.Context, collection AdminCollection) error
	UpdateAdminCollection(ctx context.Context, previous, next AdminCollection) error
	DeleteAdminCollection(ctx context.Context, collection AdminCollection) error
}

// AdminPhotoRepository owns canonical photo metadata. Public photo records are
// derived from it only while a photo is published, matching the collection
// publication model and keeping drafts private.
type AdminPhotoRepository interface {
	ListAdminPhotos(ctx context.Context) ([]AdminPhoto, error)
	ListAdminPhotosByCollection(ctx context.Context, collectionID string) ([]AdminPhoto, error)
	GetAdminPhotoByID(ctx context.Context, id string) (AdminPhoto, bool, error)
	CreateAdminPhoto(ctx context.Context, photo AdminPhoto) error
	UpdateAdminPhoto(ctx context.Context, previous, next AdminPhoto) error
	ReorderAdminPhotos(ctx context.Context, previous, next []AdminPhoto) error
}

// AdminPhotoCollectionGuardRepository provides the stronger write path used
// when a photo is public. It keeps a photo write conditional on the current
// published collection revision so archive and publication cannot interleave.
type AdminPhotoCollectionGuardRepository interface {
	UpdateAdminPhotoForPublishedCollection(ctx context.Context, previous, next AdminPhoto, collection AdminCollection) error
}

type MemoryRepository struct {
	collections       []Collection
	collectionsBySlug map[string]Collection
	photos            []Photo
	photosByID        map[string]Photo
	adminPhotos       []AdminPhoto
	adminPhotosByID   map[string]AdminPhoto
	adminCollections  []AdminCollection
	adminByID         map[string]AdminCollection
}

func NewMemoryRepository(collections []Collection, photos []Photo) *MemoryRepository {
	repository := &MemoryRepository{
		collections:       append([]Collection(nil), collections...),
		collectionsBySlug: make(map[string]Collection, len(collections)),
		photos:            append([]Photo(nil), photos...),
		photosByID:        make(map[string]Photo, len(photos)),
		adminPhotos:       make([]AdminPhoto, 0, len(photos)),
		adminPhotosByID:   make(map[string]AdminPhoto, len(photos)),
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
		adminPhoto := AdminPhoto{
			Photo:            photo,
			Status:           PublicationPublished,
			ProcessingStatus: ProcessingNotRequired,
			AltText:          photo.Title,
			Tags:             []string{},
			Version:          1,
		}
		repository.adminPhotos = append(repository.adminPhotos, adminPhoto)
		repository.adminPhotosByID[photo.ID] = adminPhoto
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

func (repository *MemoryRepository) ListAdminPhotosByCollection(_ context.Context, collectionID string) ([]AdminPhoto, error) {
	photos := make([]AdminPhoto, 0)
	for _, photo := range repository.adminPhotos {
		if photo.CollectionID == collectionID {
			photos = append(photos, photo)
		}
	}
	return photos, nil
}

func (repository *MemoryRepository) ListAdminPhotos(_ context.Context) ([]AdminPhoto, error) {
	photos := append([]AdminPhoto(nil), repository.adminPhotos...)
	sort.Slice(photos, func(left, right int) bool {
		if photos[left].CollectionID == photos[right].CollectionID {
			return photos[left].Order < photos[right].Order
		}
		return photos[left].CollectionID < photos[right].CollectionID
	})
	return photos, nil
}

func (repository *MemoryRepository) GetAdminPhotoByID(_ context.Context, id string) (AdminPhoto, bool, error) {
	photo, found := repository.adminPhotosByID[id]
	return photo, found, nil
}

func (repository *MemoryRepository) CreateAdminPhoto(_ context.Context, photo AdminPhoto) error {
	if _, found := repository.adminPhotosByID[photo.ID]; found {
		return ErrAlreadyExists
	}
	repository.adminPhotos = append(repository.adminPhotos, photo)
	repository.adminPhotosByID[photo.ID] = photo
	return nil
}

func (repository *MemoryRepository) UpdateAdminPhoto(_ context.Context, previous, next AdminPhoto) error {
	current, found := repository.adminPhotosByID[previous.ID]
	if !found || current.Version != previous.Version {
		return ErrVersionConflict
	}

	repository.replaceAdminPhoto(previous.ID, next)
	repository.reconcilePublicPhoto(current, next)
	return nil
}

func (repository *MemoryRepository) UpdateAdminPhotoForPublishedCollection(ctx context.Context, previous, next AdminPhoto, collection AdminCollection) error {
	currentCollection, found := repository.adminByID[collection.ID]
	if !found || currentCollection.Status != PublicationPublished || currentCollection.Version != collection.Version {
		return ErrVersionConflict
	}
	return repository.UpdateAdminPhoto(ctx, previous, next)
}

func (repository *MemoryRepository) ReorderAdminPhotos(_ context.Context, previous, next []AdminPhoto) error {
	if len(previous) != len(next) {
		return ErrVersionConflict
	}
	for index, photo := range previous {
		current, found := repository.adminPhotosByID[photo.ID]
		if !found || current.Version != photo.Version || next[index].ID != photo.ID {
			return ErrVersionConflict
		}
	}
	for _, photo := range next {
		current := repository.adminPhotosByID[photo.ID]
		repository.replaceAdminPhoto(photo.ID, photo)
		repository.reconcilePublicPhoto(current, photo)
	}
	return nil
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

	wasPublished := current.Status == PublicationPublished
	isPublished := next.Status == PublicationPublished
	switch {
	case wasPublished && !isPublished:
		repository.removePublicCollection(current.Collection)
	case !wasPublished && isPublished:
		repository.collections = append(repository.collections, next.Collection)
		sort.Slice(repository.collections, func(left, right int) bool {
			return repository.collections[left].Order < repository.collections[right].Order
		})
		repository.collectionsBySlug[next.Slug] = next.Collection
	case wasPublished && isPublished:
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

func (repository *MemoryRepository) DeleteAdminCollection(_ context.Context, collection AdminCollection) error {
	current, found := repository.adminByID[collection.ID]
	if !found || current.Version != collection.Version {
		return ErrVersionConflict
	}

	for index, candidate := range repository.adminCollections {
		if candidate.ID == collection.ID {
			repository.adminCollections = append(repository.adminCollections[:index], repository.adminCollections[index+1:]...)
			break
		}
	}
	delete(repository.adminByID, collection.ID)
	return nil
}

func (repository *MemoryRepository) removePublicCollection(collection Collection) {
	for index, current := range repository.collections {
		if current.ID == collection.ID {
			repository.collections = append(repository.collections[:index], repository.collections[index+1:]...)
			break
		}
	}
	delete(repository.collectionsBySlug, collection.Slug)
}

func (repository *MemoryRepository) replaceAdminPhoto(id string, next AdminPhoto) {
	for index, photo := range repository.adminPhotos {
		if photo.ID == id {
			repository.adminPhotos[index] = next
			break
		}
	}
	repository.adminPhotosByID[id] = next
}

func (repository *MemoryRepository) reconcilePublicPhoto(previous, next AdminPhoto) {
	wasPublished := previous.Status == PublicationPublished
	isPublished := next.Status == PublicationPublished
	switch {
	case wasPublished && !isPublished:
		repository.removePublicPhoto(previous.Photo)
	case !wasPublished && isPublished:
		repository.photos = append(repository.photos, next.Photo)
		repository.photosByID[next.ID] = next.Photo
	case wasPublished && isPublished:
		for index, photo := range repository.photos {
			if photo.ID == previous.ID {
				repository.photos[index] = next.Photo
				break
			}
		}
		repository.photosByID[next.ID] = next.Photo
	}
	sort.Slice(repository.photos, func(left, right int) bool {
		if repository.photos[left].CollectionID == repository.photos[right].CollectionID {
			return repository.photos[left].Order < repository.photos[right].Order
		}
		return repository.photos[left].CollectionID < repository.photos[right].CollectionID
	})
}

func (repository *MemoryRepository) removePublicPhoto(photo Photo) {
	for index, current := range repository.photos {
		if current.ID == photo.ID {
			repository.photos = append(repository.photos[:index], repository.photos[index+1:]...)
			break
		}
	}
	delete(repository.photosByID, photo.ID)
}
