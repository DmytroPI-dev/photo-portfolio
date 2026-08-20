data "aws_iam_policy_document" "gallery_api_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "gallery_api" {
  name               = "${local.name_prefix}-gallery-api-lambda"
  assume_role_policy = data.aws_iam_policy_document.gallery_api_assume_role.json
}

resource "aws_iam_role_policy_attachment" "gallery_api_basic_execution" {
  role       = aws_iam_role.gallery_api.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "gallery_api_metadata_read" {
  statement {
    effect = "Allow"

    # Public routes issue only GetItem and Query operations. Write access is
    # reserved for the separately run metadata bootstrap/admin path.
    actions = [
      "dynamodb:GetItem",
      "dynamodb:Query",
    ]

    resources = [aws_dynamodb_table.gallery_metadata.arn]
  }
}

resource "aws_iam_role_policy" "gallery_api_metadata_read" {
  name   = "${local.name_prefix}-gallery-metadata-read"
  role   = aws_iam_role.gallery_api.id
  policy = data.aws_iam_policy_document.gallery_api_metadata_read.json
}

# Collection edits are applied with one TransactWriteItems request so the
# canonical record, private index, and any published public copies stay in
# sync. DynamoDB also authorizes the PutItem and DeleteItem operations carried
# by that transaction. Publishing a photo also condition-checks its target
# collection, so the role needs that DynamoDB action as well. No standalone
# UpdateItem or broad table access is used.
data "aws_iam_policy_document" "gallery_api_metadata_write" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:ConditionCheckItem",
      "dynamodb:DeleteItem",
      "dynamodb:PutItem",
      "dynamodb:TransactWriteItems",
    ]
    resources = [aws_dynamodb_table.gallery_metadata.arn]
  }
}

resource "aws_iam_role_policy" "gallery_api_metadata_write" {
  name   = "${local.name_prefix}-gallery-metadata-write"
  role   = aws_iam_role.gallery_api.id
  policy = data.aws_iam_policy_document.gallery_api_metadata_write.json
}

# The API signs object operations but never receives original image bytes. The
# resource scope limits it to objects in this portfolio's private bucket.
data "aws_iam_policy_document" "gallery_api_originals" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = ["${aws_s3_bucket.gallery_originals.arn}/*"]
  }
}

resource "aws_iam_role_policy" "gallery_api_originals" {
  name   = "${local.name_prefix}-gallery-originals"
  role   = aws_iam_role.gallery_api.id
  policy = data.aws_iam_policy_document.gallery_api_originals.json
}

# Admin retry requests may enqueue only an existing private original for the
# dedicated worker. The API cannot inspect or consume queued jobs.
data "aws_iam_policy_document" "gallery_api_processing_queue" {
  statement {
    effect    = "Allow"
    actions   = ["sqs:SendMessage"]
    resources = [aws_sqs_queue.gallery_image_processing.arn]
  }
}

resource "aws_iam_role_policy" "gallery_api_processing_queue" {
  name   = "${local.name_prefix}-gallery-processing-queue"
  role   = aws_iam_role.gallery_api.id
  policy = data.aws_iam_policy_document.gallery_api_processing_queue.json
}

# Permanent deletion removes only this portfolio's private originals and
# derivative outputs. ListBucket is constrained to the derivative bucket so
# the API can clean an entire versioned photo prefix after a partial render.
data "aws_iam_policy_document" "gallery_api_derivative_cleanup" {
  statement {
    effect    = "Allow"
    actions   = ["s3:DeleteObject"]
    resources = ["${aws_s3_bucket.gallery_originals.arn}/*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.gallery_derivatives.arn]
  }

  statement {
    effect    = "Allow"
    actions   = ["s3:DeleteObject"]
    resources = ["${aws_s3_bucket.gallery_derivatives.arn}/derivatives/*"]
  }


  # No wildcard CloudFront permission is granted while delivery is deferred.
  # Once a distribution exists, the API may invalidate only that distribution.
  dynamic "statement" {
    for_each = local.media_distribution_enabled ? [aws_cloudfront_distribution.gallery_media[0].arn] : []
    content {
      effect    = "Allow"
      actions   = ["cloudfront:CreateInvalidation"]
      resources = [statement.value]
    }
  }
}

resource "aws_iam_role_policy" "gallery_api_derivative_cleanup" {
  name   = "${local.name_prefix}-gallery-derivative-cleanup"
  role   = aws_iam_role.gallery_api.id
  policy = data.aws_iam_policy_document.gallery_api_derivative_cleanup.json
}

# The worker has its own role, so a compromised image processor cannot mint
# browser upload URLs or operate the API's Cognito-protected HTTP surface.
data "aws_iam_policy_document" "gallery_image_worker_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "gallery_image_worker" {
  name               = "${local.name_prefix}-image-worker-lambda"
  assume_role_policy = data.aws_iam_policy_document.gallery_image_worker_assume_role.json
}

resource "aws_iam_role_policy_attachment" "gallery_image_worker_basic_execution" {
  role       = aws_iam_role.gallery_image_worker.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "gallery_image_worker_queue" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:ChangeMessageVisibility",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:ReceiveMessage",
    ]
    resources = [aws_sqs_queue.gallery_image_processing.arn]
  }
}

resource "aws_iam_role_policy" "gallery_image_worker_queue" {
  name   = "${local.name_prefix}-image-worker-queue"
  role   = aws_iam_role.gallery_image_worker.id
  policy = data.aws_iam_policy_document.gallery_image_worker_queue.json
}

# Processing uses the same conditional canonical-photo updates as the admin
# API. The narrow set mirrors its DynamoDB permissions instead of granting a
# broad table policy.
data "aws_iam_policy_document" "gallery_image_worker_metadata" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:DeleteItem",
      "dynamodb:PutItem",
      "dynamodb:TransactWriteItems",
    ]
    resources = [aws_dynamodb_table.gallery_metadata.arn]
  }
}

resource "aws_iam_role_policy" "gallery_image_worker_metadata" {
  name   = "${local.name_prefix}-image-worker-metadata"
  role   = aws_iam_role.gallery_image_worker.id
  policy = data.aws_iam_policy_document.gallery_image_worker_metadata.json
}

data "aws_iam_policy_document" "gallery_image_worker_objects" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObjectTagging",
    ]
    resources = ["${aws_s3_bucket.gallery_originals.arn}/originals/*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.gallery_derivatives.arn}/derivatives/*"]
  }
}

resource "aws_iam_role_policy" "gallery_image_worker_objects" {
  name   = "${local.name_prefix}-image-worker-objects"
  role   = aws_iam_role.gallery_image_worker.id
  policy = data.aws_iam_policy_document.gallery_image_worker_objects.json
}
