package lambdaadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/httpapi"
	"github.com/aws/aws-lambda-go/events"
)

func TestAdapterProxiesAPIGatewayV2Request(t *testing.T) {
	handler := New(httpapi.NewHandler(gallery.NewSeedRepository()))
	event := eventFromJSON(t, `{
        "rawPath": "/collections/nature",
        "headers": {"origin": "http://localhost:5173"},
        "requestContext": {"http": {"method": "GET", "path": "/collections/nature"}}
    }`)

	response, err := handler(context.Background(), event)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if response.Headers["Access-Control-Allow-Origin"] != "http://localhost:5173" {
		t.Fatalf("CORS header = %q, want local Vite origin", response.Headers["Access-Control-Allow-Origin"])
	}

	var body gallery.CollectionDetail
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != "nature" || len(body.Photos) != 5 {
		t.Fatalf("collection response = %#v, want Nature with five photos", body)
	}
}

func TestAdapterKeepsNotFoundJSONResponse(t *testing.T) {
	handler := New(httpapi.NewHandler(gallery.NewSeedRepository()))
	event := eventFromJSON(t, `{
        "rawPath": "/photos/missing",
        "requestContext": {"http": {"method": "GET", "path": "/photos/missing"}}
    }`)

	response, err := handler(context.Background(), event)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
	if response.Headers["Content-Type"] != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want JSON", response.Headers["Content-Type"])
	}
}

func eventFromJSON(t *testing.T, raw string) events.APIGatewayV2HTTPRequest {
	t.Helper()

	var event events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("decode API Gateway event: %v", err)
	}
	return event
}
