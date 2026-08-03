package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
)

type fakeOriginalStore struct{}

func (fakeOriginalStore) PresignPut(_ context.Context, key, _ string) (string, error) {
	return "https://uploads.example.test/" + key, nil
}

func (fakeOriginalStore) PresignGet(_ context.Context, key string) (string, error) {
	return "https://previews.example.test/" + key, nil
}

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

func TestAdminReadsExposeTheExpectedCollectionAndPhotoShapes(t *testing.T) {
	collectionsResponse := request(t, http.MethodGet, "/admin/collections")
	if collectionsResponse.Code != http.StatusOK {
		t.Fatalf("collection status = %d, want %d", collectionsResponse.Code, http.StatusOK)
	}

	var collections struct {
		Items []gallery.Collection `json:"items"`
	}
	decodeJSON(t, collectionsResponse, &collections)
	if len(collections.Items) != 3 || collections.Items[0].ID != "drawings" {
		t.Fatalf("admin collections = %#v, want seeded collections", collections.Items)
	}

	photosResponse := request(t, http.MethodGet, "/admin/photos")
	if photosResponse.Code != http.StatusOK {
		t.Fatalf("photo status = %d, want %d", photosResponse.Code, http.StatusOK)
	}

	var photos struct {
		Items []gallery.Photo `json:"items"`
	}
	decodeJSON(t, photosResponse, &photos)
	if len(photos.Items) != 16 || photos.Items[0].ID != "drawing-01" || photos.Items[15].ID != "travel-05" {
		t.Fatalf("admin photos = %#v, want all seeded photos in display order", photos.Items)
	}
}

func TestAdminPhotoDetailUsesItsOwnRoutePrefix(t *testing.T) {
	response := request(t, http.MethodGet, "/admin/photos/drawing-01")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var photo gallery.Photo
	decodeJSON(t, response, &photo)
	if photo.ID != "drawing-01" || photo.Title != "Stillness" {
		t.Fatalf("photo = %#v, want drawing-01 Stillness", photo)
	}
}

func TestAdminCollectionsCreateDraftAndEditWithVersion(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())
	create := httptest.NewRequest(http.MethodPost, "/admin/collections", strings.NewReader(`{"slug":"sketches","title":"Sketches","description":"First studies.","order":4}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)

	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}

	var collection gallery.AdminCollection
	decodeJSON(t, created, &collection)
	if collection.Status != gallery.PublicationDraft || collection.Version != 1 || collection.ID != "sketches" {
		t.Fatalf("created collection = %#v, want draft sketches version 1", collection)
	}

	update := httptest.NewRequest(http.MethodPatch, "/admin/collections/sketches", strings.NewReader(`{"title":"Field Sketches","version":1}`))
	update.Header.Set("Content-Type", "application/json")
	updated := httptest.NewRecorder()
	handler.ServeHTTP(updated, update)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body = %s", updated.Code, http.StatusOK, updated.Body.String())
	}
	decodeJSON(t, updated, &collection)
	if collection.Title != "Field Sketches" || collection.Version != 2 {
		t.Fatalf("updated collection = %#v, want title and version 2", collection)
	}
}

func TestAdminCollectionPublishesDraftAndArchivesEmptyCollection(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())

	created := requestWithHandler(t, handler, http.MethodPost, "/admin/collections", `{"slug":"sketches","title":"Sketches","order":4}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}

	published := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/publish", `{"version":1}`)
	if published.Code != http.StatusOK {
		t.Fatalf("publish status = %d, want %d; body = %s", published.Code, http.StatusOK, published.Body.String())
	}
	var collection gallery.AdminCollection
	decodeJSON(t, published, &collection)
	if collection.Status != gallery.PublicationPublished || collection.Version != 2 {
		t.Fatalf("published collection = %#v, want published version 2", collection)
	}

	publicDetail := requestWithHandler(t, handler, http.MethodGet, "/collections/sketches", "")
	if publicDetail.Code != http.StatusOK {
		t.Fatalf("public detail status = %d, want %d; body = %s", publicDetail.Code, http.StatusOK, publicDetail.Body.String())
	}

	archived := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/archive", `{"version":2}`)
	if archived.Code != http.StatusOK {
		t.Fatalf("archive status = %d, want %d; body = %s", archived.Code, http.StatusOK, archived.Body.String())
	}
	decodeJSON(t, archived, &collection)
	if collection.Status != gallery.PublicationArchived || collection.Version != 3 {
		t.Fatalf("archived collection = %#v, want archived version 3", collection)
	}

	publicDetail = requestWithHandler(t, handler, http.MethodGet, "/collections/sketches", "")
	if publicDetail.Code != http.StatusNotFound {
		t.Fatalf("archived public detail status = %d, want %d", publicDetail.Code, http.StatusNotFound)
	}
}

func TestAdminCollectionCannotArchiveWithPublishedPhotos(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())
	response := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/drawings/archive", `{"version":1}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("archive status = %d, want %d; body = %s", response.Code, http.StatusConflict, response.Body.String())
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, response, &body)
	if body.Error.Code != "collection_has_published_photos" {
		t.Fatalf("error code = %q, want collection_has_published_photos", body.Error.Code)
	}
}

func TestAdminPhotoCRUDAndLifecycleReconcilesPublicCopies(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())

	created := requestWithHandler(t, handler, http.MethodPost, "/admin/photos", `{
        "id":"fixture-photo","title":"Fixture Photo","description":"For console testing.","src":"/images/1.jpg",
        "collectionId":"drawings","width":1350,"height":1800,"order":20,"altText":"Fixture image",
        "tags":["fixture"," drawing "],"focalPointX":0.5,"focalPointY":0.5
    }`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create photo status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}
	var photo gallery.AdminPhoto
	decodeJSON(t, created, &photo)
	if photo.Status != gallery.PublicationDraft || photo.Version != 1 || photo.ID != "fixture-photo" || len(photo.Tags) != 2 {
		t.Fatalf("created photo = %#v, want draft fixture photo", photo)
	}

	publicDetail := requestWithHandler(t, handler, http.MethodGet, "/photos/fixture-photo", "")
	if publicDetail.Code != http.StatusNotFound {
		t.Fatalf("draft public detail status = %d, want %d", publicDetail.Code, http.StatusNotFound)
	}

	published := requestWithHandler(t, handler, http.MethodPost, "/admin/photos/fixture-photo/publish", `{"version":1}`)
	if published.Code != http.StatusOK {
		t.Fatalf("publish photo status = %d, want %d; body = %s", published.Code, http.StatusOK, published.Body.String())
	}
	decodeJSON(t, published, &photo)
	if photo.Status != gallery.PublicationPublished || photo.Version != 2 {
		t.Fatalf("published photo = %#v, want published version 2", photo)
	}

	publicDetail = requestWithHandler(t, handler, http.MethodGet, "/photos/fixture-photo", "")
	if publicDetail.Code != http.StatusOK {
		t.Fatalf("published public detail status = %d, want %d; body = %s", publicDetail.Code, http.StatusOK, publicDetail.Body.String())
	}

	updated := requestWithHandler(t, handler, http.MethodPatch, "/admin/photos/fixture-photo", `{"title":"Moved Fixture","collectionId":"travel","order":30,"version":2}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("update photo status = %d, want %d; body = %s", updated.Code, http.StatusOK, updated.Body.String())
	}
	decodeJSON(t, updated, &photo)
	if photo.Title != "Moved Fixture" || photo.CollectionID != "travel" || photo.Version != 3 {
		t.Fatalf("updated photo = %#v, want moved published photo version 3", photo)
	}

	archived := requestWithHandler(t, handler, http.MethodPost, "/admin/photos/fixture-photo/archive", `{"version":3}`)
	if archived.Code != http.StatusOK {
		t.Fatalf("archive photo status = %d, want %d; body = %s", archived.Code, http.StatusOK, archived.Body.String())
	}
	publicDetail = requestWithHandler(t, handler, http.MethodGet, "/photos/fixture-photo", "")
	if publicDetail.Code != http.StatusNotFound {
		t.Fatalf("archived public detail status = %d, want %d", publicDetail.Code, http.StatusNotFound)
	}

	readOnly := requestWithHandler(t, handler, http.MethodPatch, "/admin/photos/fixture-photo", `{"title":"Nope","version":4}`)
	if readOnly.Code != http.StatusConflict {
		t.Fatalf("update archived photo status = %d, want %d; body = %s", readOnly.Code, http.StatusConflict, readOnly.Body.String())
	}
	assertErrorCode(t, readOnly, "archived_photo_read_only")

	restored := requestWithHandler(t, handler, http.MethodPost, "/admin/photos/fixture-photo/restore", `{"version":4}`)
	if restored.Code != http.StatusOK {
		t.Fatalf("restore photo status = %d, want %d; body = %s", restored.Code, http.StatusOK, restored.Body.String())
	}
	decodeJSON(t, restored, &photo)
	if photo.Status != gallery.PublicationDraft || photo.Version != 5 {
		t.Fatalf("restored photo = %#v, want draft version 5", photo)
	}
}

func TestAdminImageUploadCreatesPrivateDraftWithPreview(t *testing.T) {
	handler := NewHandlerWithOriginals(gallery.NewSeedRepository(), fakeOriginalStore{})

	prepared := requestWithHandler(t, handler, http.MethodPost, "/admin/uploads", `{"filename":"winter-study.JPG","contentType":"image/jpeg","size":2048}`)
	if prepared.Code != http.StatusCreated {
		t.Fatalf("prepare upload status = %d, want %d; body = %s", prepared.Code, http.StatusCreated, prepared.Body.String())
	}
	var upload createUploadResponse
	decodeJSON(t, prepared, &upload)
	if !strings.HasPrefix(upload.PhotoID, "photo-") || !strings.HasPrefix(upload.OriginalKey, "originals/"+upload.PhotoID+"/") || upload.UploadURL == "" {
		t.Fatalf("upload preparation = %#v, want generated private upload details", upload)
	}

	created := requestWithHandler(t, handler, http.MethodPost, "/admin/photos", fmt.Sprintf(`{
        "uploadId":%q,"originalKey":%q,"title":"Winter Study","collectionId":"drawings",
        "width":1200,"height":1600,"altText":"Winter study","tags":["study"]
    }`, upload.PhotoID, upload.OriginalKey))
	if created.Code != http.StatusCreated {
		t.Fatalf("create uploaded photo status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}
	var photo gallery.AdminPhoto
	decodeJSON(t, created, &photo)
	if photo.ID != upload.PhotoID || photo.OriginalKey != upload.OriginalKey || photo.ProcessingStatus != gallery.ProcessingPending || photo.Order != 7 {
		t.Fatalf("created uploaded photo = %#v, want private pending draft appended to drawings", photo)
	}

	preview := requestWithHandler(t, handler, http.MethodGet, "/admin/photos/"+photo.ID+"/preview", "")
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want %d; body = %s", preview.Code, http.StatusOK, preview.Body.String())
	}
	var previewBody map[string]string
	decodeJSON(t, preview, &previewBody)
	if previewBody["url"] != "https://previews.example.test/"+upload.OriginalKey {
		t.Fatalf("preview URL = %q, want signed private original", previewBody["url"])
	}

	publish := requestWithHandler(t, handler, http.MethodPost, "/admin/photos/"+photo.ID+"/publish", `{"version":1}`)
	if publish.Code != http.StatusConflict {
		t.Fatalf("publish pending photo status = %d, want %d; body = %s", publish.Code, http.StatusConflict, publish.Body.String())
	}
	assertErrorCode(t, publish, "image_not_ready")
}

func TestAdminPhotoReorderUpdatesPublicCollectionOrder(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())
	reordered := requestWithHandler(t, handler, http.MethodPost, "/admin/photos/reorder", `{
        "collectionId":"drawings",
        "photos":[
            {"id":"drawing-02","version":1},
            {"id":"drawing-01","version":1},
            {"id":"drawing-03","version":1},
            {"id":"drawing-04","version":1},
            {"id":"drawing-05","version":1},
            {"id":"drawing-06","version":1}
        ]
    }`)
	if reordered.Code != http.StatusOK {
		t.Fatalf("reorder status = %d, want %d; body = %s", reordered.Code, http.StatusOK, reordered.Body.String())
	}

	publicCollection := requestWithHandler(t, handler, http.MethodGet, "/collections/drawings", "")
	if publicCollection.Code != http.StatusOK {
		t.Fatalf("public collection status = %d, want %d", publicCollection.Code, http.StatusOK)
	}
	var detail gallery.CollectionDetail
	decodeJSON(t, publicCollection, &detail)
	if detail.Photos[0].ID != "drawing-02" || detail.Photos[1].ID != "drawing-01" {
		t.Fatalf("reordered public photos = %#v, want drawing-02 then drawing-01", detail.Photos[:2])
	}
}

func TestArchivedCollectionIsReadOnlyAndCanBePermanentlyDeletedWithConfirmation(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())

	created := requestWithHandler(t, handler, http.MethodPost, "/admin/collections", `{"slug":"sketches","title":"Sketches","order":4}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}

	archived := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/archive", `{"version":1}`)
	if archived.Code != http.StatusOK {
		t.Fatalf("archive status = %d, want %d; body = %s", archived.Code, http.StatusOK, archived.Body.String())
	}

	update := requestWithHandler(t, handler, http.MethodPatch, "/admin/collections/sketches", `{"title":"Changed after archive","version":2}`)
	if update.Code != http.StatusConflict {
		t.Fatalf("update archived status = %d, want %d; body = %s", update.Code, http.StatusConflict, update.Body.String())
	}
	assertErrorCode(t, update, "archived_collection_read_only")

	restored := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/restore", `{"version":2}`)
	if restored.Code != http.StatusOK {
		t.Fatalf("restore status = %d, want %d; body = %s", restored.Code, http.StatusOK, restored.Body.String())
	}
	var restoredCollection gallery.AdminCollection
	decodeJSON(t, restored, &restoredCollection)
	if restoredCollection.Status != gallery.PublicationDraft || restoredCollection.Version != 3 {
		t.Fatalf("restored collection = %#v, want draft version 3", restoredCollection)
	}

	publicDetail := requestWithHandler(t, handler, http.MethodGet, "/collections/sketches", "")
	if publicDetail.Code != http.StatusNotFound {
		t.Fatalf("restored draft public detail status = %d, want %d", publicDetail.Code, http.StatusNotFound)
	}

	restoreDraft := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/restore", `{"version":3}`)
	if restoreDraft.Code != http.StatusConflict {
		t.Fatalf("restore draft status = %d, want %d; body = %s", restoreDraft.Code, http.StatusConflict, restoreDraft.Body.String())
	}
	assertErrorCode(t, restoreDraft, "invalid_state")

	archivedAgain := requestWithHandler(t, handler, http.MethodPost, "/admin/collections/sketches/archive", `{"version":3}`)
	if archivedAgain.Code != http.StatusOK {
		t.Fatalf("archive restored draft status = %d, want %d; body = %s", archivedAgain.Code, http.StatusOK, archivedAgain.Body.String())
	}

	notArchived := requestWithHandler(t, handler, http.MethodDelete, "/admin/collections/drawings", `{"version":1,"confirmationSlug":"drawings"}`)
	if notArchived.Code != http.StatusConflict {
		t.Fatalf("delete published status = %d, want %d; body = %s", notArchived.Code, http.StatusConflict, notArchived.Body.String())
	}
	assertErrorCode(t, notArchived, "collection_not_archived")

	wrongConfirmation := requestWithHandler(t, handler, http.MethodDelete, "/admin/collections/sketches", `{"version":4,"confirmationSlug":"wrong"}`)
	if wrongConfirmation.Code != http.StatusBadRequest {
		t.Fatalf("delete wrong confirmation status = %d, want %d; body = %s", wrongConfirmation.Code, http.StatusBadRequest, wrongConfirmation.Body.String())
	}
	assertErrorCode(t, wrongConfirmation, "invalid_confirmation")

	deleted := requestWithHandler(t, handler, http.MethodDelete, "/admin/collections/sketches", `{"version":4,"confirmationSlug":"sketches"}`)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d; body = %s", deleted.Code, http.StatusNoContent, deleted.Body.String())
	}

	detail := requestWithHandler(t, handler, http.MethodGet, "/admin/collections/sketches", "")
	if detail.Code != http.StatusNotFound {
		t.Fatalf("deleted admin detail status = %d, want %d", detail.Code, http.StatusNotFound)
	}
}

func TestAdminCollectionRejectsMultipleJSONValues(t *testing.T) {
	handler := NewHandler(gallery.NewSeedRepository())
	request := httptest.NewRequest(http.MethodPost, "/admin/collections", strings.NewReader("{\"slug\":\"sketches\",\"title\":\"Sketches\",\"order\":4}\n{\"slug\":\"ignored\"}"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestUnknownResourceReturnsJSONNotFound(t *testing.T) {
	response := request(t, http.MethodGet, "/collections/missing")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	assertErrorCode(t, response, "not_found")
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
	if headers := response.Header().Get("Access-Control-Allow-Headers"); headers != "Authorization, Content-Type" {
		t.Fatalf("CORS headers = %q, want Authorization and Content-Type", headers)
	}
	if methods := response.Header().Get("Access-Control-Allow-Methods"); methods != "GET, POST, PATCH, DELETE, OPTIONS" {
		t.Fatalf("CORS methods = %q, want GET, POST, PATCH, DELETE, OPTIONS", methods)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, response, &body)
	if body.Error.Code != want {
		t.Fatalf("error code = %q, want %q", body.Error.Code, want)
	}
}

func request(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewHandler(gallery.NewSeedRepository())
	return requestWithHandler(t, handler, method, path, "")
}

func requestWithHandler(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
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
