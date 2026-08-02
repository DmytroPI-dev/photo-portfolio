data "aws_caller_identity" "current" {}

locals {
  name_prefix = "${var.project}-${var.environment}"

  api_function_name = "${local.name_prefix}-gallery-api"
  api_name          = "${local.name_prefix}-api"
  metadata_table    = "${local.name_prefix}-gallery-metadata"
  admin_user_pool   = "${local.name_prefix}-gallery-admin"

  # The account ID makes Cognito's globally shared hosted-UI prefix predictable
  # without borrowing any name from the separate Mandelbrot project.
  admin_cognito_domain = "${local.name_prefix}-admin-${data.aws_caller_identity.current.account_id}"

  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
    Repository  = "photo-portfolio"
  }
}
