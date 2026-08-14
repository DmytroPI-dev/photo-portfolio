# Image Worker Container

This Go Lambda container consumes the image-processing SQS queue and calls the
`vips` binary installed in its Alpine base image. It creates `thumbnail`
(480px), `medium` (1200px), and `large` (2400px) WebP derivatives beneath a
deterministic `derivatives/{photo-id}/v1/` prefix.

The first Terraform apply deliberately creates the ECR repository, buckets,
queue, and policies without creating the Lambda. After that apply, build and
push an ARM64 image, set `image_worker_image_uri` to the resulting immutable
digest in the ignored Terraform variables file, then plan/apply again. Do not
use a mutable image tag for this value.

The source original receives `gallery-processing=complete` only after all
variants and metadata are written successfully. Its S3 lifecycle rule then
expires it 30 days after original object creation.
