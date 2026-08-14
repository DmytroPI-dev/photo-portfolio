# Originals are private by design. The browser uploads directly with a narrow,
# short-lived URL minted by the API Lambda; only future processed derivatives
# will be exposed through the media CloudFront distribution.
resource "aws_s3_bucket" "gallery_originals" {
  bucket = "${local.name_prefix}-gallery-originals-${data.aws_caller_identity.current.account_id}"

  tags = local.tags
}

resource "aws_s3_bucket_public_access_block" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 evaluates this CORS rule for the direct POST, independently of API
# Gateway's CORS configuration. Keep it limited to the admin origins.
resource "aws_s3_bucket_cors_configuration" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  cors_rule {
    allowed_headers = ["content-type"]
    allowed_methods = ["POST"]
    allowed_origins = var.admin_upload_cors_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }
}

# Derivatives stay private until the later media CloudFront distribution is in
# place. Separating them from originals lets the worker have write-only access
# to the public-facing output and keeps upload inputs out of the delivery path.
resource "aws_s3_bucket" "gallery_derivatives" {
  bucket = "${local.name_prefix}-gallery-derivatives-${data.aws_caller_identity.current.account_id}"

  tags = local.tags
}

resource "aws_s3_bucket_public_access_block" "gallery_derivatives" {
  bucket = aws_s3_bucket.gallery_derivatives.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "gallery_derivatives" {
  bucket = aws_s3_bucket.gallery_derivatives.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "gallery_derivatives" {
  bucket = aws_s3_bucket.gallery_derivatives.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 measures lifecycle expiration from object creation, not from when this tag
# is added. Processing normally takes minutes, so a successfully processed
# original is retained for roughly 30 days after its browser upload.
resource "aws_s3_bucket_lifecycle_configuration" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  rule {
    id     = "expire-successfully-processed-originals"
    status = "Enabled"

    filter {
      tag {
        key   = "gallery-processing"
        value = "complete"
      }
    }

    expiration {
      days = 30
    }
  }
}

# The worker consumes S3 notifications from SQS rather than directly from S3.
# This gives failures durable retries and a DLQ without coupling S3 delivery to
# a single Lambda invocation.
resource "aws_s3_bucket_notification" "gallery_originals" {
  bucket = aws_s3_bucket.gallery_originals.id

  queue {
    queue_arn     = aws_sqs_queue.gallery_image_processing.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix = "originals/"
  }

  depends_on = [aws_sqs_queue_policy.gallery_image_processing]
}
