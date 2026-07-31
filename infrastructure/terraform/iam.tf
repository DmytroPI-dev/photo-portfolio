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

# The first deployed Lambda uses in-memory metadata. DynamoDB permissions are
# intentionally added with the DynamoDB repository, rather than granted early.
