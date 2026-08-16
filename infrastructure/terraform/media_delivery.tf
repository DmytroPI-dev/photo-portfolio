# The certificate is created independently from the distribution. This lets the
# developer add ACM's validation CNAME to Cloudflare, wait for issuance, and
# only then opt into the CloudFront resources with media_certificate_arn.
resource "aws_acm_certificate" "gallery_media" {
  provider          = aws.us_east_1
  domain_name       = var.media_domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# A distribution is conditional because CloudFront rejects a pending ACM
# certificate. The public hostname is a distinct media-only origin; neither the
# Azure-hosted gallery nor the locally developed admin console changes here.
resource "aws_cloudfront_origin_access_control" "gallery_media" {
  count = local.media_distribution_enabled ? 1 : 0

  name                              = "${local.name_prefix}-gallery-media"
  description                       = "Signed CloudFront reads from private gallery derivatives."
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# Public derivatives are safe to load cross-origin because they are not
# authenticated or personalized. The header is required for Three.js/WebGL to
# upload an Azure-hosted gallery image into a GPU texture without tainting it.
resource "aws_cloudfront_response_headers_policy" "gallery_media_cors" {
  count = local.media_distribution_enabled ? 1 : 0

  name    = "${local.name_prefix}-gallery-media-cors"
  comment = "Permit credential-free cross-origin WebGL texture loads."

  cors_config {
    access_control_allow_credentials = false
    origin_override                  = true

    access_control_allow_headers {
      items = ["*"]
    }

    access_control_allow_methods {
      items = ["GET", "HEAD", "OPTIONS"]
    }

    access_control_allow_origins {
      items = ["*"]
    }
  }
}

resource "aws_cloudfront_distribution" "gallery_media" {
  count = local.media_distribution_enabled ? 1 : 0

  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Public gallery derivative delivery from a private S3 origin."
  aliases             = [var.media_domain_name]
  default_root_object = ""
  price_class         = "PriceClass_All"

  origin {
    domain_name              = aws_s3_bucket.gallery_derivatives.bucket_regional_domain_name
    origin_id                = "gallery-derivatives-s3"
    origin_access_control_id = aws_cloudfront_origin_access_control.gallery_media[0].id
  }

  # Derivative keys include the processing profile version, so replacing an
  # image creates a new URL instead of needing a CDN invalidation.
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "gallery-derivatives-s3"
    compress         = true

    viewer_protocol_policy     = "redirect-to-https"
    response_headers_policy_id = aws_cloudfront_response_headers_policy.gallery_media_cors[0].id
    min_ttl                    = 0
    default_ttl                = 86400
    max_ttl                    = 31536000

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = var.media_certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}

# Public access remains blocked at the bucket. This narrow policy authorizes
# only signed reads sent by this exact media distribution to immutable outputs.
data "aws_iam_policy_document" "gallery_derivatives_cloudfront" {
  count = local.media_distribution_enabled ? 1 : 0

  statement {
    sid    = "AllowCloudFrontReadOnly"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.gallery_derivatives.arn}/derivatives/*"]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.gallery_media[0].arn]
    }
  }
}

resource "aws_s3_bucket_policy" "gallery_derivatives_cloudfront" {
  count = local.media_distribution_enabled ? 1 : 0

  bucket = aws_s3_bucket.gallery_derivatives.id
  policy = data.aws_iam_policy_document.gallery_derivatives_cloudfront[0].json
}
