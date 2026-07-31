variable "aws_profile" {
  description = "Local AWS CLI profile used by Terraform. CI will use OIDC instead."
  type        = string
  default     = "default"
}

variable "aws_region" {
  description = "Primary AWS region for the API and DynamoDB table."
  type        = string
  default     = "eu-central-1"
}

variable "project" {
  description = "Project name used for resource naming and tags."
  type        = string
  default     = "photo-portfolio"
}

variable "environment" {
  description = "Deployment environment name."
  type        = string
  default     = "prod"
}

variable "lambda_package_path" {
  description = "Path to the ARM64 Lambda zip produced by scripts/package-gallery-api.sh."
  type        = string
  default     = "../../services/gallery-api/build/gallery-api.zip"
}

variable "api_memory_size" {
  description = "Memory assigned to the metadata API Lambda in megabytes."
  type        = number
  default     = 256
}

variable "api_timeout_seconds" {
  description = "Maximum execution time for the metadata API Lambda."
  type        = number
  default     = 10
}

variable "log_retention_days" {
  description = "CloudWatch log retention for the API Lambda and HTTP API."
  type        = number
  default     = 14
}

variable "allowed_cors_origins" {
  description = "Browser origins permitted to call the public metadata API."
  type        = list(string)
  default = [
    "http://localhost:5173",
    "http://127.0.0.1:5173",
    "https://photo-gallery.i-dmytro.org",
  ]
}
