package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

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
	handler.collections(writer, request)
}

func (handler *Handler) adminCollection(writer http.ResponseWriter, request *http.Request) {
	if !requireGet(writer, request) {
		return
	}

	slug := strings.TrimPrefix(request.URL.Path, "/admin/collections/")
	if slug == "" || strings.Contains(slug, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}

	handler.writeCollectionDetail(writer, request, slug)
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
	handler.photo(writer, request)
}

func (handler *Handler) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if _, allowed := handler.allowedOrigins[origin]; allowed {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}

		if request.Method == http.MethodOptions {
			writer.Header().Set("Access-Control-Allow-Methods", http.MethodGet+", OPTIONS")
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

	writer.Header().Set("Allow", http.MethodGet+", OPTIONS")
	writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is supported")
	return false
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
