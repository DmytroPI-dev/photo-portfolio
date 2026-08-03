package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/storage"
)

type Handler struct {
	repository     gallery.Repository
	originals      storage.Presigner
	allowedOrigins map[string]struct{}
}

func NewHandler(repository gallery.Repository) http.Handler {
	return NewHandlerWithOriginals(repository, nil)
}

// NewHandlerWithOriginals adds the private-originals capability used in AWS.
// Passing nil deliberately leaves the local metadata-only server usable while
// returning a clear response for upload-only routes.
func NewHandlerWithOriginals(repository gallery.Repository, originals storage.Presigner) http.Handler {
	handler := &Handler{
		repository: repository,
		originals:  originals,
		// API Gateway owns production CORS later. These local Vite origins make
		// the service immediately useful without turning every local response
		// into a permissive wildcard policy.
		allowedOrigins: map[string]struct{}{
			"http://localhost:5173": {},
			"http://127.0.0.1:5173": {},
			"http://localhost:5174": {},
			"http://127.0.0.1:5174": {},
		},
	}

	return handler.withCORS(http.HandlerFunc(handler.serveHTTP))
}

func (handler *Handler) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.URL.Path == "/health":
		handler.health(writer, request)
	case request.URL.Path == "/collections":
		handler.collections(writer, request)
	case strings.HasPrefix(request.URL.Path, "/collections/"):
		handler.collection(writer, request)
	case strings.HasPrefix(request.URL.Path, "/photos/"):
		handler.photo(writer, request)
	case request.URL.Path == "/admin/collections":
		handler.adminCollections(writer, request)
	case strings.HasPrefix(request.URL.Path, "/admin/collections/"):
		handler.adminCollection(writer, request)
	case request.URL.Path == "/admin/uploads":
		handler.createUpload(writer, request)
	case request.URL.Path == "/admin/photos":
		handler.adminPhotos(writer, request)
	case request.URL.Path == "/admin/photos/reorder":
		handler.reorderAdminPhotos(writer, request)
	case strings.HasPrefix(request.URL.Path, "/admin/photos/"):
		handler.adminPhoto(writer, request)
	default:
		writeError(writer, http.StatusNotFound, "not_found", "route not found")
	}
}

func (handler *Handler) health(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) collections(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	collections, err := handler.repository.ListCollections(request.Context())
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{"items": collections})
}

func (handler *Handler) collection(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	slug := strings.TrimPrefix(request.URL.Path, "/collections/")
	if slug == "" || strings.Contains(slug, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}

	handler.writeCollectionDetail(writer, request, slug)
}

func (handler *Handler) writeCollectionDetail(writer http.ResponseWriter, request *http.Request, slug string) {
	collection, found, err := handler.repository.GetCollectionBySlug(request.Context(), slug)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}

	photos, err := handler.repository.ListPhotosByCollection(request.Context(), collection.ID)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, gallery.CollectionDetail{
		Collection: collection,
		Photos:     photos,
	})
}

func (handler *Handler) photo(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	id := strings.TrimPrefix(request.URL.Path, "/photos/")
	if id == "" || strings.Contains(id, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}

	handler.writePhoto(writer, request, id)
}

// writePhoto keeps public and protected detail routes on the same repository
// lookup path. Route-specific handlers parse their own prefixes before calling
// it, avoiding accidental treatment of /admin/photos/{id} as a public URL.
func (handler *Handler) writePhoto(writer http.ResponseWriter, request *http.Request, id string) {
	photo, found, err := handler.repository.GetPhotoByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}

	writeJSON(writer, http.StatusOK, photo)
}

// API Gateway's JWT authorizer protects these routes in AWS. The local HTTP
// server deliberately does not impersonate Cognito, letting handler tests stay
// focused on request/response behavior. Never expose this handler directly on
// a public network outside API Gateway.
func (handler *Handler) adminCollections(writer http.ResponseWriter, request *http.Request) {
	store, ok := handler.repository.(gallery.AdminCollectionRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return
	}

	switch request.Method {
	case http.MethodGet:
		collections, err := store.ListAdminCollections(request.Context())
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"items": collections})
	case http.MethodPost:
		handler.createAdminCollection(writer, request, store)
	default:
		writeMethodNotAllowed(writer, http.MethodGet+", "+http.MethodPost+", OPTIONS")
	}
}

func (handler *Handler) adminCollection(writer http.ResponseWriter, request *http.Request) {
	store, ok := handler.repository.(gallery.AdminCollectionRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return
	}

	path := strings.TrimPrefix(request.URL.Path, "/admin/collections/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" || len(parts) > 2 {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	id := parts[0]
	if len(parts) == 2 {
		if request.Method != http.MethodPost {
			writeMethodNotAllowed(writer, http.MethodPost+", OPTIONS")
			return
		}
		switch parts[1] {
		case "publish":
			handler.transitionAdminCollection(writer, request, store, id, gallery.PublicationPublished)
		case "archive":
			handler.transitionAdminCollection(writer, request, store, id, gallery.PublicationArchived)
		case "restore":
			handler.transitionAdminCollection(writer, request, store, id, gallery.PublicationDraft)
		default:
			writeError(writer, http.StatusNotFound, "not_found", "collection action not found")
		}
		return
	}

	switch request.Method {
	case http.MethodGet:
		collection, found, err := store.GetAdminCollectionByID(request.Context(), id)
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		if !found {
			writeError(writer, http.StatusNotFound, "not_found", "collection not found")
			return
		}
		writeJSON(writer, http.StatusOK, collection)
	case http.MethodPatch:
		handler.updateAdminCollection(writer, request, store, id)
	case http.MethodDelete:
		handler.deleteAdminCollection(writer, request, store, id)
	default:
		writeMethodNotAllowed(writer, http.MethodGet+", "+http.MethodPatch+", "+http.MethodDelete+", OPTIONS")
	}
}

type createCollectionRequest struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CoverPhotoID string `json:"coverPhotoId"`
	Order        int    `json:"order"`
}

type updateCollectionRequest struct {
	Slug         *string `json:"slug"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	CoverPhotoID *string `json:"coverPhotoId"`
	Order        *int    `json:"order"`
	Version      *int    `json:"version"`
}

// Collection state transitions intentionally accept only the version. Metadata
// belongs on PATCH, while publishing and archiving have separate, auditable
// domain actions with their own safety checks.
type collectionTransitionRequest struct {
	Version int `json:"version"`
}

// Collection deletion is intentionally more constrained than archiving. The
// exact slug avoids accidental confirmation, and the handler also verifies the
// canonical record has no remaining photo ownership before removing metadata.
type deleteCollectionRequest struct {
	Version          int    `json:"version"`
	ConfirmationSlug string `json:"confirmationSlug"`
}

func (handler *Handler) createAdminCollection(writer http.ResponseWriter, request *http.Request, store gallery.AdminCollectionRepository) {
	var input createCollectionRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}

	input.Slug = strings.TrimSpace(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	if !validSlug(input.Slug) {
		writeError(writer, http.StatusBadRequest, "invalid_slug", "slug must use lowercase letters, numbers, and single hyphens")
		return
	}
	if input.Title == "" || input.Order < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_collection", "title and a positive display order are required")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	collection := gallery.AdminCollection{
		Collection: gallery.Collection{
			ID:           input.Slug,
			Slug:         input.Slug,
			Title:        input.Title,
			Description:  input.Description,
			CoverPhotoID: input.CoverPhotoID,
			Order:        input.Order,
		},
		Status:    gallery.PublicationDraft,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.CreateAdminCollection(request.Context(), collection); err != nil {
		writeAdminStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, collection)
}

func (handler *Handler) updateAdminCollection(writer http.ResponseWriter, request *http.Request, store gallery.AdminCollectionRepository, id string) {
	var input updateCollectionRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}

	current, found, err := store.GetAdminCollectionByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	if current.Status == gallery.PublicationArchived {
		writeError(writer, http.StatusConflict, "archived_collection_read_only", "archived collections cannot be edited")
		return
	}
	// The first admin-console rollout used a disabled form field for version,
	// and some browsers submit that as zero. A positive version still enforces
	// optimistic locking; an omitted or zero legacy value uses this just-read
	// canonical version and remains protected by the DynamoDB transaction.
	if input.Version != nil && *input.Version > 0 && *input.Version != current.Version {
		writeError(writer, http.StatusConflict, "version_conflict", "collection was changed by another request; reload and try again")
		return
	}
	if input.Slug != nil && strings.TrimSpace(*input.Slug) != current.Slug {
		// The initial public layout uses the slug as the DynamoDB collection
		// partition. Keep it immutable until an explicit migration exists.
		writeError(writer, http.StatusBadRequest, "immutable_slug", "collection slug cannot be changed")
		return
	}

	next := current
	if input.Title != nil {
		next.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		next.Description = *input.Description
	}
	if input.CoverPhotoID != nil {
		next.CoverPhotoID = *input.CoverPhotoID
	}
	if input.Order != nil {
		next.Order = *input.Order
	}
	if next.Title == "" || next.Order < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_collection", "title and a positive display order are required")
		return
	}
	next.Version++
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := store.UpdateAdminCollection(request.Context(), current, next); err != nil {
		writeAdminStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, next)
}

func (handler *Handler) transitionAdminCollection(writer http.ResponseWriter, request *http.Request, store gallery.AdminCollectionRepository, id string, target gallery.PublicationStatus) {
	var input collectionTransitionRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	if input.Version < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_version", "a positive collection version is required")
		return
	}

	current, found, err := store.GetAdminCollectionByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	if input.Version != current.Version {
		writeError(writer, http.StatusConflict, "version_conflict", "collection was changed by another request; reload and try again")
		return
	}
	if current.Status == target {
		if target == gallery.PublicationDraft {
			writeError(writer, http.StatusConflict, "invalid_state", "only archived collections can be restored")
			return
		}
		writeJSON(writer, http.StatusOK, current)
		return
	}

	switch target {
	case gallery.PublicationPublished:
		if current.Status != gallery.PublicationDraft {
			writeError(writer, http.StatusConflict, "invalid_state", "only draft collections can be published")
			return
		}
	case gallery.PublicationDraft:
		if current.Status != gallery.PublicationArchived {
			writeError(writer, http.StatusConflict, "invalid_state", "only archived collections can be restored")
			return
		}
	case gallery.PublicationArchived:
		if current.Status != gallery.PublicationDraft && current.Status != gallery.PublicationPublished {
			writeError(writer, http.StatusConflict, "invalid_state", "collection is already archived")
			return
		}
		if current.Status == gallery.PublicationPublished {
			photos, err := store.ListAdminPhotosByCollection(request.Context(), current.ID)
			if err != nil {
				writeRepositoryError(writer, err)
				return
			}
			for _, photo := range photos {
				if photo.Status != gallery.PublicationPublished {
					continue
				}
				writeError(writer, http.StatusConflict, "collection_has_published_photos", "archive or move all published photos before archiving this collection")
				return
			}
		}
	}

	next := current
	next.Status = target
	next.Version++
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.UpdateAdminCollection(request.Context(), current, next); err != nil {
		writeAdminStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, next)
}

func (handler *Handler) deleteAdminCollection(writer http.ResponseWriter, request *http.Request, store gallery.AdminCollectionRepository, id string) {
	var input deleteCollectionRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	if input.Version < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_version", "a positive collection version is required")
		return
	}

	current, found, err := store.GetAdminCollectionByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	if input.Version != current.Version {
		writeError(writer, http.StatusConflict, "version_conflict", "collection was changed by another request; reload and try again")
		return
	}
	if current.Status != gallery.PublicationArchived {
		writeError(writer, http.StatusConflict, "collection_not_archived", "only archived collections can be permanently deleted")
		return
	}
	if input.ConfirmationSlug != current.Slug {
		writeError(writer, http.StatusBadRequest, "invalid_confirmation", "type the exact collection slug to confirm deletion")
		return
	}

	photos, err := store.ListAdminPhotosByCollection(request.Context(), current.ID)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if len(photos) > 0 {
		writeError(writer, http.StatusConflict, "collection_has_photos", "move or delete all photos before permanently deleting this collection")
		return
	}

	if err := store.DeleteAdminCollection(request.Context(), current); err != nil {
		writeAdminStoreError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) adminPhotos(writer http.ResponseWriter, request *http.Request) {
	store, ok := handler.repository.(gallery.AdminPhotoRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return
	}

	switch request.Method {
	case http.MethodGet:
		photos, err := store.ListAdminPhotos(request.Context())
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"items": photos})
	case http.MethodPost:
		handler.createAdminPhoto(writer, request, store)
	default:
		writeMethodNotAllowed(writer, http.MethodGet+", "+http.MethodPost+", OPTIONS")
	}
}

func (handler *Handler) adminPhoto(writer http.ResponseWriter, request *http.Request) {
	store, ok := handler.repository.(gallery.AdminPhotoRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return
	}

	path := strings.TrimPrefix(request.URL.Path, "/admin/photos/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" || len(parts) > 2 {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}
	id := parts[0]
	if len(parts) == 2 {
		if parts[1] == "preview" {
			if request.Method != http.MethodGet {
				writeMethodNotAllowed(writer, http.MethodGet+", OPTIONS")
				return
			}
			handler.adminPhotoPreview(writer, request, store, id)
			return
		}
		if request.Method != http.MethodPost {
			writeMethodNotAllowed(writer, http.MethodPost+", OPTIONS")
			return
		}
		switch parts[1] {
		case "publish":
			handler.transitionAdminPhoto(writer, request, store, id, gallery.PublicationPublished)
		case "archive":
			handler.transitionAdminPhoto(writer, request, store, id, gallery.PublicationArchived)
		case "restore":
			handler.transitionAdminPhoto(writer, request, store, id, gallery.PublicationDraft)
		default:
			writeError(writer, http.StatusNotFound, "not_found", "photo action not found")
		}
		return
	}

	switch request.Method {
	case http.MethodGet:
		photo, found, err := store.GetAdminPhotoByID(request.Context(), id)
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		if !found {
			writeError(writer, http.StatusNotFound, "not_found", "photo not found")
			return
		}
		writeJSON(writer, http.StatusOK, photo)
	case http.MethodPatch:
		handler.updateAdminPhoto(writer, request, store, id)
	default:
		writeMethodNotAllowed(writer, http.MethodGet+", "+http.MethodPatch+", OPTIONS")
	}
}

type createPhotoRequest struct {
	// UploadID and OriginalKey are issued by POST /admin/uploads. They are not
	// user-editable values in the console, but retaining the legacy fields lets
	// existing seeded/test metadata use the same create route.
	UploadID     string   `json:"uploadId"`
	OriginalKey  string   `json:"originalKey"`
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Src          string   `json:"src"`
	CollectionID string   `json:"collectionId"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Year         string   `json:"year"`
	Location     string   `json:"location"`
	Featured     bool     `json:"featured"`
	Order        int      `json:"order"`
	AltText      string   `json:"altText"`
	Tags         []string `json:"tags"`
	FocalPointX  float64  `json:"focalPointX"`
	FocalPointY  float64  `json:"focalPointY"`
}

type updatePhotoRequest struct {
	Title        *string   `json:"title"`
	Description  *string   `json:"description"`
	Src          *string   `json:"src"`
	CollectionID *string   `json:"collectionId"`
	Width        *int      `json:"width"`
	Height       *int      `json:"height"`
	Year         *string   `json:"year"`
	Location     *string   `json:"location"`
	Featured     *bool     `json:"featured"`
	Order        *int      `json:"order"`
	AltText      *string   `json:"altText"`
	Tags         *[]string `json:"tags"`
	FocalPointX  *float64  `json:"focalPointX"`
	FocalPointY  *float64  `json:"focalPointY"`
	Version      *int      `json:"version"`
}

type photoTransitionRequest struct {
	Version int `json:"version"`
}

const maxOriginalUploadBytes int64 = 25 << 20

type createUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

type createUploadResponse struct {
	PhotoID     string `json:"photoId"`
	OriginalKey string `json:"originalKey"`
	UploadURL   string `json:"uploadUrl"`
}

func (handler *Handler) createUpload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeMethodNotAllowed(writer, http.MethodPost+", OPTIONS")
		return
	}
	if handler.originals == nil {
		writeError(writer, http.StatusNotImplemented, "uploads_not_configured", "original image uploads are not configured")
		return
	}

	var input createUploadRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	contentType, extension, valid := validOriginalUpload(input.ContentType, input.Filename)
	if !valid || input.Size < 1 || input.Size > maxOriginalUploadBytes {
		writeError(writer, http.StatusBadRequest, "invalid_upload", "choose a JPEG, PNG, or WebP image up to 25 MB")
		return
	}

	photoID, err := generatedPhotoID()
	if err != nil {
		log.Printf("generate upload ID: %v", err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "unable to prepare image upload")
		return
	}
	originalKey := "originals/" + photoID + "/original" + extension
	uploadURL, err := handler.originals.PresignPut(request.Context(), originalKey, contentType)
	if err != nil {
		log.Printf("presign upload for %q: %v", photoID, err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "unable to prepare image upload")
		return
	}
	writeJSON(writer, http.StatusCreated, createUploadResponse{PhotoID: photoID, OriginalKey: originalKey, UploadURL: uploadURL})
}

func (handler *Handler) adminPhotoPreview(writer http.ResponseWriter, request *http.Request, store gallery.AdminPhotoRepository, id string) {
	if handler.originals == nil {
		writeError(writer, http.StatusNotImplemented, "uploads_not_configured", "private image previews are not configured")
		return
	}
	photo, found, err := store.GetAdminPhotoByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}
	if photo.OriginalKey == "" {
		writeError(writer, http.StatusNotFound, "preview_not_found", "this photo has no private original")
		return
	}
	previewURL, err := handler.originals.PresignGet(request.Context(), photo.OriginalKey)
	if err != nil {
		log.Printf("presign preview for %q: %v", id, err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "unable to prepare image preview")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"url": previewURL})
}

type reorderPhotoRequest struct {
	CollectionID string `json:"collectionId"`
	Photos       []struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
	} `json:"photos"`
}

func (handler *Handler) createAdminPhoto(writer http.ResponseWriter, request *http.Request, store gallery.AdminPhotoRepository) {
	var input createPhotoRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	input.UploadID = strings.TrimSpace(input.UploadID)
	input.OriginalKey = strings.TrimSpace(input.OriginalKey)
	input.ID = strings.TrimSpace(input.ID)
	usesUpload := input.UploadID != "" || input.OriginalKey != ""
	photoID := input.ID
	if usesUpload {
		if !validSlug(input.UploadID) || !validOriginalKey(input.UploadID, input.OriginalKey) {
			writeError(writer, http.StatusBadRequest, "invalid_upload", "a valid prepared image upload is required")
			return
		}
		photoID = input.UploadID
	}
	if !validSlug(photoID) {
		writeError(writer, http.StatusBadRequest, "invalid_id", "photo ID must use lowercase letters, numbers, and single hyphens")
		return
	}

	// TODO: FocalPointX and FocalPointY are intentionally defaulted to 0.5 (centre) until derivative
	// generation and a visual crop editor are implemented; at that point, wire them from input.FocalPointX/Y.
	photo := gallery.AdminPhoto{Photo: gallery.Photo{
		ID:           photoID,
		Title:        strings.TrimSpace(input.Title),
		Description:  input.Description,
		Src:          strings.TrimSpace(input.Src),
		CollectionID: strings.TrimSpace(input.CollectionID),
		Width:        input.Width,
		Height:       input.Height,
		Year:         strings.TrimSpace(input.Year),
		Location:     strings.TrimSpace(input.Location),
		Featured:     input.Featured,
		Order:        input.Order,
	}, ProcessingStatus: gallery.ProcessingNotRequired, AltText: strings.TrimSpace(input.AltText), Tags: normalizeTags(input.Tags), FocalPointX: 0.5, FocalPointY: 0.5}
	if usesUpload {
		photo.OriginalKey = input.OriginalKey
		photo.ProcessingStatus = gallery.ProcessingPending
	}
	if photo.Order == 0 {
		existing, err := store.ListAdminPhotosByCollection(request.Context(), photo.CollectionID)
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		for _, existingPhoto := range existing {
			if existingPhoto.Order >= photo.Order {
				photo.Order = existingPhoto.Order + 1
			}
		}
	}
	if !handler.validateAdminPhoto(writer, request, photo, false) {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	photo.Status = gallery.PublicationDraft
	photo.Version = 1
	photo.CreatedAt = now
	photo.UpdatedAt = now
	if err := store.CreateAdminPhoto(request.Context(), photo); err != nil {
		writePhotoStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, photo)
}

func (handler *Handler) updateAdminPhoto(writer http.ResponseWriter, request *http.Request, store gallery.AdminPhotoRepository, id string) {
	var input updatePhotoRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}

	current, found, err := store.GetAdminPhotoByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}
	if current.Status == gallery.PublicationArchived {
		writeError(writer, http.StatusConflict, "archived_photo_read_only", "archived photos cannot be edited")
		return
	}
	if input.Version != nil && *input.Version > 0 && *input.Version != current.Version {
		writeError(writer, http.StatusConflict, "version_conflict", "photo was changed by another request; reload and try again")
		return
	}

	next := applyPhotoUpdate(current, input)
	if !handler.validateAdminPhoto(writer, request, next, current.Status == gallery.PublicationPublished) {
		return
	}
	next.Version++
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.UpdateAdminPhoto(request.Context(), current, next); err != nil {
		writePhotoStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, next)
}

func (handler *Handler) transitionAdminPhoto(writer http.ResponseWriter, request *http.Request, store gallery.AdminPhotoRepository, id string, target gallery.PublicationStatus) {
	var input photoTransitionRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	if input.Version < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_version", "a positive photo version is required")
		return
	}

	current, found, err := store.GetAdminPhotoByID(request.Context(), id)
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}
	if input.Version != current.Version {
		writeError(writer, http.StatusConflict, "version_conflict", "photo was changed by another request; reload and try again")
		return
	}
	if current.Status == target {
		if target == gallery.PublicationDraft {
			writeError(writer, http.StatusConflict, "invalid_state", "only archived photos can be restored")
			return
		}
		writeJSON(writer, http.StatusOK, current)
		return
	}

	switch target {
	case gallery.PublicationPublished:
		if current.Status != gallery.PublicationDraft {
			writeError(writer, http.StatusConflict, "invalid_state", "only draft photos can be published")
			return
		}
		if !handler.validateAdminPhoto(writer, request, current, true) {
			return
		}
	case gallery.PublicationDraft:
		if current.Status != gallery.PublicationArchived {
			writeError(writer, http.StatusConflict, "invalid_state", "only archived photos can be restored")
			return
		}
	case gallery.PublicationArchived:
		if current.Status != gallery.PublicationDraft && current.Status != gallery.PublicationPublished {
			writeError(writer, http.StatusConflict, "invalid_state", "photo is already archived")
			return
		}
	}

	next := current
	next.Status = target
	next.Version++
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.UpdateAdminPhoto(request.Context(), current, next); err != nil {
		writePhotoStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, next)
}

func (handler *Handler) reorderAdminPhotos(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeMethodNotAllowed(writer, http.MethodPost+", OPTIONS")
		return
	}
	store, ok := handler.repository.(gallery.AdminPhotoRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return
	}
	var input reorderPhotoRequest
	if !decodeRequestJSON(writer, request, &input) {
		return
	}
	input.CollectionID = strings.TrimSpace(input.CollectionID)
	if input.CollectionID == "" || len(input.Photos) == 0 || len(input.Photos) > 16 {
		writeError(writer, http.StatusBadRequest, "invalid_reorder", "a collection and between one and sixteen photos are required")
		return
	}

	previous := make([]gallery.AdminPhoto, 0, len(input.Photos))
	next := make([]gallery.AdminPhoto, 0, len(input.Photos))
	seen := make(map[string]struct{}, len(input.Photos))
	for index, item := range input.Photos {
		if item.ID == "" || item.Version < 1 {
			writeError(writer, http.StatusBadRequest, "invalid_reorder", "each reordered photo needs an ID and positive version")
			return
		}
		if _, duplicate := seen[item.ID]; duplicate {
			writeError(writer, http.StatusBadRequest, "invalid_reorder", "each photo can appear only once in a reorder")
			return
		}
		seen[item.ID] = struct{}{}
		current, found, err := store.GetAdminPhotoByID(request.Context(), item.ID)
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		if !found || current.CollectionID != input.CollectionID || current.Version != item.Version {
			writeError(writer, http.StatusConflict, "version_conflict", "photo order changed by another request; reload and try again")
			return
		}
		previous = append(previous, current)
		current.Order = index + 1
		current.Version++
		current.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		next = append(next, current)
	}
	if err := store.ReorderAdminPhotos(request.Context(), previous, next); err != nil {
		writePhotoStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": next})
}

func applyPhotoUpdate(current gallery.AdminPhoto, input updatePhotoRequest) gallery.AdminPhoto {
	next := current
	if input.Title != nil {
		next.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		next.Description = *input.Description
	}
	if input.Src != nil {
		next.Src = strings.TrimSpace(*input.Src)
	}
	if input.CollectionID != nil {
		next.CollectionID = strings.TrimSpace(*input.CollectionID)
	}
	if input.Width != nil {
		next.Width = *input.Width
	}
	if input.Height != nil {
		next.Height = *input.Height
	}
	if input.Year != nil {
		next.Year = strings.TrimSpace(*input.Year)
	}
	if input.Location != nil {
		next.Location = strings.TrimSpace(*input.Location)
	}
	if input.Featured != nil {
		next.Featured = *input.Featured
	}
	if input.Order != nil {
		next.Order = *input.Order
	}
	if input.AltText != nil {
		next.AltText = strings.TrimSpace(*input.AltText)
	}
	if input.Tags != nil {
		next.Tags = normalizeTags(*input.Tags)
	}
	if input.FocalPointX != nil {
		next.FocalPointX = *input.FocalPointX
	}
	if input.FocalPointY != nil {
		next.FocalPointY = *input.FocalPointY
	}
	return next
}

func (handler *Handler) validateAdminPhoto(writer http.ResponseWriter, request *http.Request, photo gallery.AdminPhoto, publishing bool) bool {
	if photo.Title == "" || (photo.Src == "" && photo.OriginalKey == "") || photo.CollectionID == "" || photo.Width < 1 || photo.Height < 1 || photo.Order < 1 || photo.FocalPointX < 0 || photo.FocalPointX > 1 || photo.FocalPointY < 0 || photo.FocalPointY > 1 {
		writeError(writer, http.StatusBadRequest, "invalid_photo", "title, image, collection, and positive dimensions/order are required")
		return false
	}
	if publishing && (photo.Src == "" || photo.ProcessingStatus != gallery.ProcessingReady && photo.ProcessingStatus != gallery.ProcessingNotRequired) {
		writeError(writer, http.StatusConflict, "image_not_ready", "wait for image processing before publishing this photo")
		return false
	}
	collections, ok := handler.repository.(gallery.AdminCollectionRepository)
	if !ok {
		writeError(writer, http.StatusNotImplemented, "admin_not_configured", "administrator metadata is not configured")
		return false
	}
	collection, found, err := collections.GetAdminCollectionByID(request.Context(), photo.CollectionID)
	if err != nil {
		writeRepositoryError(writer, err)
		return false
	}
	if !found || collection.Status == gallery.PublicationArchived {
		writeError(writer, http.StatusBadRequest, "invalid_collection", "choose an existing draft or published collection")
		return false
	}
	if publishing && collection.Status != gallery.PublicationPublished {
		writeError(writer, http.StatusConflict, "collection_not_published", "publish the collection before publishing this photo")
		return false
	}
	return true
}

func generatedPhotoID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "photo-" + hex.EncodeToString(bytes), nil
}

func validOriginalUpload(contentType, filename string) (string, string, bool) {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	lastDot := strings.LastIndex(filename, ".")
	if lastDot < 1 || lastDot == len(filename)-1 {
		return "", "", false
	}
	extension := strings.ToLower(filename[lastDot:])
	allowed := map[string]struct {
		contentType string
		extension   string
	}{
		"image/jpeg": {contentType: "image/jpeg", extension: ".jpg"},
		"image/png":  {contentType: "image/png", extension: ".png"},
		"image/webp": {contentType: "image/webp", extension: ".webp"},
	}
	value, found := allowed[contentType]
	if !found || extension == "." || extension == "" {
		return "", "", false
	}
	return value.contentType, value.extension, true
}

func validOriginalKey(photoID, originalKey string) bool {
	prefix := "originals/" + photoID + "/original"
	if !strings.HasPrefix(originalKey, prefix) {
		return false
	}
	extension := strings.TrimPrefix(originalKey, prefix)
	return extension == ".jpg" || extension == ".png" || extension == ".webp"
}

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		value := strings.TrimSpace(tag)
		if value == "" {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func (handler *Handler) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if _, allowed := handler.allowedOrigins[origin]; allowed {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}

		if request.Method == http.MethodOptions {
			writer.Header().Set("Access-Control-Allow-Methods", http.MethodGet+", "+http.MethodPost+", "+http.MethodPatch+", "+http.MethodDelete+", OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func requireGet(writer http.ResponseWriter, request *http.Request) bool {
	if request.Method == http.MethodGet {
		return true
	}

	writeMethodNotAllowed(writer, http.MethodGet+", OPTIONS")
	return false
}

func writeMethodNotAllowed(writer http.ResponseWriter, allowed string) {
	writer.Header().Set("Allow", allowed)
	writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported for this route")
}

func decodeRequestJSON(writer http.ResponseWriter, request *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, 32<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "invalid_json", "request body must contain exactly one JSON value")
		return false
	}
	return true
}

func validSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 80 || slug[0] == '-' || slug[len(slug)-1] == '-' {
		return false
	}

	previousHyphen := false
	for _, character := range slug {
		isLowercaseLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if isLowercaseLetter || isDigit {
			previousHyphen = false
			continue
		}
		if character == '-' && !previousHyphen {
			previousHyphen = true
			continue
		}
		return false
	}
	return true
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func writeRepositoryError(writer http.ResponseWriter, err error) {
	// Keep the underlying DynamoDB error in Lambda logs, but never expose table
	// names, request details, or AWS errors through the public API response.
	log.Printf("gallery repository request failed: %v", err)
	writeError(writer, http.StatusInternalServerError, "internal_error", "unable to read gallery data")
}

func writeAdminStoreError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gallery.ErrAlreadyExists):
		writeError(writer, http.StatusConflict, "already_exists", "a collection with that identifier already exists")
	case errors.Is(err, gallery.ErrVersionConflict):
		writeError(writer, http.StatusConflict, "version_conflict", "collection was changed by another request; reload and try again")
	default:
		log.Printf("gallery admin write failed: %v", err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "unable to update gallery metadata")
	}
}

func writePhotoStoreError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gallery.ErrAlreadyExists):
		writeError(writer, http.StatusConflict, "already_exists", "a photo with that identifier already exists")
	case errors.Is(err, gallery.ErrVersionConflict):
		writeError(writer, http.StatusConflict, "version_conflict", "photo was changed by another request; reload and try again")
	default:
		log.Printf("gallery admin photo write failed: %v", err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "unable to update photo metadata")
	}
}
