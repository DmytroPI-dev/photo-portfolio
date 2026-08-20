resource "aws_cloudwatch_log_group" "gallery_api" {
  name              = "/aws/lambda/${local.api_function_name}"
  retention_in_days = var.log_retention_days
}

resource "aws_lambda_function" "gallery_api" {
  function_name = local.api_function_name
  description   = "Public metadata API for the photo portfolio."

  role    = aws_iam_role.gallery_api.arn
  runtime = "provided.al2023"
  handler = "bootstrap"

  filename         = var.lambda_package_path
  source_code_hash = try(filebase64sha256(var.lambda_package_path), null)

  architectures = ["arm64"]
  memory_size   = var.api_memory_size
  timeout       = var.api_timeout_seconds

  # An empty value selects the local in-memory seed repository. Terraform sets
  # this in Lambda so production fails clearly rather than serving stale data.
  environment {
    variables = {
      GALLERY_METADATA_TABLE        = aws_dynamodb_table.gallery_metadata.name
      GALLERY_ORIGINALS_BUCKET      = aws_s3_bucket.gallery_originals.bucket
      GALLERY_PROCESSING_QUEUE_URL  = aws_sqs_queue.gallery_image_processing.url
      GALLERY_DERIVATIVES_BUCKET    = aws_s3_bucket.gallery_derivatives.bucket
      GALLERY_MEDIA_DISTRIBUTION_ID = try(aws_cloudfront_distribution.gallery_media[0].id, "")
      # The API leaves uploaded drafts unpublished until this is configured.
      # It never treats the private S3 derivative key as a browser URL.
      GALLERY_MEDIA_BASE_URL = local.media_base_url
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.gallery_api,
    aws_iam_role_policy_attachment.gallery_api_basic_execution,
    aws_iam_role_policy.gallery_api_metadata_read,
    aws_iam_role_policy.gallery_api_metadata_write,
    aws_iam_role_policy.gallery_api_originals,
    aws_iam_role_policy.gallery_api_processing_queue,
    aws_iam_role_policy.gallery_api_derivative_cleanup,
  ]
}

resource "aws_ecr_repository" "gallery_image_worker" {
  name                 = "${local.name_prefix}-image-worker"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = local.tags
}

resource "aws_ecr_lifecycle_policy" "gallery_image_worker" {
  repository = aws_ecr_repository.gallery_image_worker.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep the five newest worker images for rollback."
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 5
        }
        action = { type = "expire" }
      }
    ]
  })
}

# A first Terraform apply creates the durable storage and ECR repository. The
# function is intentionally omitted until a locally built, immutable ARM64
# image digest is supplied through image_worker_image_uri; this avoids baking a
# mutable tag or a placeholder image into the deployment.
resource "aws_cloudwatch_log_group" "gallery_image_worker" {
  count = var.image_worker_image_uri == "" ? 0 : 1

  name              = "/aws/lambda/${local.image_worker_function_name}"
  retention_in_days = var.log_retention_days
}

resource "aws_lambda_function" "gallery_image_worker" {
  count = var.image_worker_image_uri == "" ? 0 : 1

  function_name = local.image_worker_function_name
  description   = "Normalizes uploaded gallery images with libvips."

  role         = aws_iam_role.gallery_image_worker.arn
  package_type = "Image"
  image_uri    = var.image_worker_image_uri

  architectures = ["arm64"]
  memory_size   = var.image_worker_memory_size
  timeout       = var.image_worker_timeout_seconds

  ephemeral_storage {
    size = 2048
  }

  environment {
    variables = {
      GALLERY_METADATA_TABLE     = aws_dynamodb_table.gallery_metadata.name
      GALLERY_ORIGINALS_BUCKET   = aws_s3_bucket.gallery_originals.bucket
      GALLERY_DERIVATIVES_BUCKET = aws_s3_bucket.gallery_derivatives.bucket
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.gallery_image_worker,
    aws_iam_role_policy_attachment.gallery_image_worker_basic_execution,
    aws_iam_role_policy.gallery_image_worker_queue,
    aws_iam_role_policy.gallery_image_worker_metadata,
    aws_iam_role_policy.gallery_image_worker_objects,
  ]
}

resource "aws_lambda_event_source_mapping" "gallery_image_worker" {
  count = var.image_worker_image_uri == "" ? 0 : 1

  event_source_arn = aws_sqs_queue.gallery_image_processing.arn
  function_name    = aws_lambda_function.gallery_image_worker[0].arn
  batch_size       = 1
  enabled          = true

  function_response_types = ["ReportBatchItemFailures"]
}
