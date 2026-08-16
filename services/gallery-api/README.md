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
returns a 15-minute pre-signed S3 `POST` form for a JPEG, PNG, or WebP original
up to 25 MB. The form policy binds the object key and size limit; the browser
uploads directly to the encrypted private bucket, then creates a draft photo
using the returned opaque upload fields.

Camera RAW formats are intentionally unsupported. The 25 MB limit protects the
upload path; it is not a public delivery format. The deployed Go/libvips worker
creates normalized WebP derivatives and applies lifecycle cleanup to a
successful private processing input. Its S3, SQS, derivative bucket, and
container Lambda path have passed a live upload smoke test.

`GET /admin/photos/{id}/preview` returns a 10-minute authenticated preview URL
for that private original. This deployed upload flow has been smoke-tested.

The local seed-only server intentionally has no bucket configured, so those
two upload routes return `uploads_not_configured`. Public publishing remains
blocked for uploaded drafts until the image-processing worker creates a
derivative and marks the photo ready. It is then also blocked until
`GALLERY_MEDIA_BASE_URL` names the private CloudFront media distribution; the
API derives the server-managed public `Src` during the publish transition.

## Image worker

`cmd/image-worker` is an ARM64 Lambda container entry point. It receives S3
notifications through SQS, validates image dimensions before decoding, applies
EXIF-aware `vips thumbnail` transforms, and writes immutable v1 `thumbnail`
(480px), `medium` (1200px), and `large` (2400px) WebP keys. Duplicate messages
are harmless once the photo is ready; failed messages update private metadata
and remain eligible for the queue's retry/DLQ policy.

The worker infrastructure is deliberately deployed in two applies: first
create the ECR repository, private derivative bucket, SQS queue, DLQ, and S3
notification; then push an ARM64 image and supply its immutable digest as
`image_worker_image_uri` before planning the Lambda/event-source mapping. The
developer performs both Terraform applies.
