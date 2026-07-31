# Gallery API

The Gallery API provides the public collection and photo metadata contract. It
currently uses an in-memory seed repository matching the frontend placeholders;
the DynamoDB repository will replace that implementation without changing the
HTTP routes.

## Local development

```bash
GALLERY_API_ADDR=127.0.0.1:8080 go run ./cmd/api
go test ./...
```

The public routes are:

```text
GET /health
GET /collections
GET /collections/{slug}
GET /photos/{id}
```

## Lambda package

From the repository root:

```bash
./scripts/package-gallery-api.sh
```

It creates `services/gallery-api/build/gallery-api.zip`, containing the Linux
ARM64 `bootstrap` binary used by the `provided.al2023` Lambda runtime.
