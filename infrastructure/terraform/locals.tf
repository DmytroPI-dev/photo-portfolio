data "aws_caller_identity" "current" {}

locals {
  name_prefix = "${var.project}-${var.environment}"

  api_function_name = "${local.name_prefix}-gallery-api"
  api_name          = "${local.name_prefix}-api"
  metadata_table    = "${local.name_prefix}-gallery-metadata"

  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
    Repository  = "photo-portfolio"
  }
}
