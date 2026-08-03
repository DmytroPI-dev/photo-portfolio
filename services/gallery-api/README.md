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

Public gallery routes remain read-only. The deployed Lambda also serves
authenticated `/admin/*` routes, which may create and update metadata; this
command is for deterministic bootstrap or intentional resets.

## Private original uploads

When Terraform supplies `GALLERY_ORIGINALS_BUCKET`, `POST /admin/uploads`
returns a 15-minute constrained S3 POST form for a JPEG, PNG, or WebP original
up to 25 MB. S3 enforces the size limit when it receives the upload.
The browser uploads directly to that private bucket, then creates a draft photo
using the returned opaque upload fields. `GET /admin/photos/{id}/preview`
returns a 10-minute authenticated preview URL for that private original.

The local seed-only server intentionally has no bucket configured, so those
two upload routes return `uploads_not_configured`. Public publishing remains
blocked for uploaded drafts until the image-processing worker creates a
derivative and marks the photo ready.
