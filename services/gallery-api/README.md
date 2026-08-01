# Gallery API

The API exposes public collection and photo metadata. Locally it uses the
placeholder repository; the deployed Lambda reads the DynamoDB table named by
`GALLERY_METADATA_TABLE`.

## Local development

```bash
go run ./cmd/api
curl http://localhost:8080/health
```

## Bootstrap DynamoDB metadata

The command below uses the current AWS CLI/SDK identity to write the
deterministic public read models, canonical private records, and administrator
list indexes. It overwrites those fixed seed keys, so do not run it after real
administrator-authored records exist unless resetting them is intentional.

```bash
GALLERY_METADATA_TABLE=photo-portfolio-prod-gallery-metadata go run ./cmd/seed
```

The public Lambda remains read-only; this command is the only Phase 1 writer.
