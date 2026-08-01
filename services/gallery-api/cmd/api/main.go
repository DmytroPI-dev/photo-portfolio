package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/appconfig"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/httpapi"
	"github.com/DmytroPI-dev/photo-portfolio/services/gallery-api/internal/lambdaadapter"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	repository, err := appconfig.NewRepository(context.Background())
	if err != nil {
		log.Fatalf("configure gallery repository: %v", err)
	}

	handler := httpapi.NewHandler(repository)
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		// provided.al2023 invokes the custom-runtime bootstrap binary through the
		// Lambda Runtime API. The adapter preserves the local net/http handler so
		// routing and JSON responses have one source of truth.
		lambda.Start(lambdaadapter.New(handler))
		return
	}

	address := os.Getenv("GALLERY_API_ADDR")
	if address == "" {
		address = ":8080"
	}

	// The HTTP handler has no knowledge of its process or hosting environment.
	// This keeps local development behavior identical to the Lambda route logic.
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("gallery API listening on %s", address)
	log.Fatal(server.ListenAndServe())
}
