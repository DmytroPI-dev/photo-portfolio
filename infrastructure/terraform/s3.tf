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
