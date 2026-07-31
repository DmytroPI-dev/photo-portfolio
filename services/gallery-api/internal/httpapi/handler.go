package httpapi

import (
	"encoding/json"
	"net/http"
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

	writeJSON(writer, http.StatusOK, map[string]any{
		"items": handler.repository.ListCollections(),
	})
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

	collection, found := handler.repository.GetCollectionBySlug(slug)
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "collection not found")
		return
	}

	writeJSON(writer, http.StatusOK, gallery.CollectionDetail{
		Collection: collection,
		Photos:     handler.repository.ListPhotosByCollection(collection.ID),
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

	photo, found := handler.repository.GetPhotoByID(id)
	if !found {
		writeError(writer, http.StatusNotFound, "not_found", "photo not found")
		return
	}

	writeJSON(writer, http.StatusOK, photo)
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
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
