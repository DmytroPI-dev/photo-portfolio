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
