package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
)

func TestHealth(t *testing.T) {
	response := request(t, http.MethodGet, "/health")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body map[string]string
	decodeJSON(t, response, &body)
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}

func TestListCollectionsUsesDisplayOrder(t *testing.T) {
	response := request(t, http.MethodGet, "/collections")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body struct {
		Items []gallery.Collection `json:"items"`
	}
	decodeJSON(t, response, &body)

	if len(body.Items) != 3 {
		t.Fatalf("collection count = %d, want 3", len(body.Items))
	}
	if body.Items[0].Slug != "drawings" || body.Items[2].Slug != "travel" {
		t.Fatalf("collection order = %#v, want drawings through travel", body.Items)
	}
}

func TestGetCollectionIncludesOnlyItsOrderedPhotos(t *testing.T) {
	response := request(t, http.MethodGet, "/collections/nature")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body gallery.CollectionDetail
	decodeJSON(t, response, &body)

	if body.ID != "nature" {
		t.Fatalf("collection ID = %q, want nature", body.ID)
	}
	if len(body.Photos) != 5 {
		t.Fatalf("photo count = %d, want 5", len(body.Photos))
	}
	if body.Photos[0].ID != "nature-01" || body.Photos[4].ID != "nature-05" {
		t.Fatalf("unexpected photo order: %#v", body.Photos)
	}
}

func TestGetPhoto(t *testing.T) {
	response := request(t, http.MethodGet, "/photos/drawing-01")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body gallery.Photo
	decodeJSON(t, response, &body)
	if body.Title != "Stillness" || body.Src != "/images/1.jpg" {
		t.Fatalf("photo = %#v, want Stillness with its source path", body)
	}
}

func TestUnknownResourceReturnsJSONNotFound(t *testing.T) {
	response := request(t, http.MethodGet, "/collections/missing")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, response, &body)
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", body.Error.Code)
	}
}

func TestOptionsAllowsLocalViteOrigin(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())
	request := httptest.NewRequest(http.MethodOptions, "/collections", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if origin := response.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Fatalf("CORS origin = %q, want local Vite origin", origin)
	}
}

func request(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewHandler(gallery.NewSeedRepository())
	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want JSON", contentType)
	}
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}
