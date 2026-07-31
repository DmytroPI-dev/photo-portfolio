package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/gallery"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/httpapi"
)

func main() {
	address := os.Getenv("GALLERY_API_ADDR")
	if address == "" {
		address = ":8080"
	}

	// The HTTP handler has no knowledge of its process or hosting environment.
	// That keeps local development identical to the later API Gateway adapter.
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewHandler(gallery.NewSeedRepository()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("gallery API listening on %s", address)
	log.Fatal(server.ListenAndServe())
}
