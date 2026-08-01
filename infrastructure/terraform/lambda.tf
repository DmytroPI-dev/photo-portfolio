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
      GALLERY_METADATA_TABLE = aws_dynamodb_table.gallery_metadata.name
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.gallery_api,
    aws_iam_role_policy_attachment.gallery_api_basic_execution,
    aws_iam_role_policy.gallery_api_metadata_read,
  ]
}
