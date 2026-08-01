package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
)

type Handler struct {
	repository     gallery.Repository
	allowedOrigins map[string]struct{}
}

func NewHandler(repository gallery.Repository) http.Handler {
	handler := &Handler{
		repository: repository,
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
	case request.URL.Path == "/admin/photos":
		handler.adminPhotos(writer, request)
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

	id := strings.TrimPrefix(request.URL.Path, "/admin/collections/")
	if id == "" || strings.Contains(id, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
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
	default:
		writeMethodNotAllowed(writer, http.MethodGet+", "+http.MethodPatch+", OPTIONS")
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

// adminPhotos is a temporary authenticated read path over the small seeded
// portfolio. It performs one bounded collection query per collection rather
// than a DynamoDB Scan. Before drafts or larger catalogues are introduced, the
// write increment will add a dedicated private administrative photo index.
func (handler *Handler) adminPhotos(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	collections, err := handler.repository.ListCollections(request.Context())
	if err != nil {
		writeRepositoryError(writer, err)
		return
	}

	photos := make([]gallery.Photo, 0)
	for _, collection := range collections {
		collectionPhotos, err := handler.repository.ListPhotosByCollection(request.Context(), collection.ID)
		if err != nil {
			writeRepositoryError(writer, err)
			return
		}
		photos = append(photos, collectionPhotos...)
	}

	sort.SliceStable(photos, func(left, right int) bool {
		return photos[left].Order < photos[right].Order
	})
	writeJSON(writer, http.StatusOK, map[string]any{"items": photos})
}

func (handler *Handler) adminPhoto(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	id := strings.TrimPrefix(request.URL.Path, "/admin/photos/")
	if id == "" || strings.Contains(id, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}

	handler.writePhoto(writer, request, id)
}

func (handler *Handler) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if _, allowed := handler.allowedOrigins[origin]; allowed {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}

		if request.Method == http.MethodOptions {
			writer.Header().Set("Access-Control-Allow-Methods", http.MethodGet+", "+http.MethodPost+", "+http.MethodPatch+", OPTIONS")
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
