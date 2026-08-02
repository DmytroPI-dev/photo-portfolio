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
    "http://localhost:5174",
    "http://127.0.0.1:5174",
    "https://photo-gallery.i-dmytro.org",
  ]
}

variable "admin_callback_urls" {
  description = "OAuth callback URLs for the React-admin SPA. The production hostname is reserved until its CloudFront distribution is deployed."
  type        = list(string)
  default = [
    "http://localhost:5174/auth/callback",
    "http://127.0.0.1:5174/auth/callback",
    "https://admin.photo-gallery.i-dmytro.org/auth/callback",
  ]
}

variable "admin_logout_urls" {
  description = "OAuth logout return URLs for the React-admin SPA."
  type        = list(string)
  default = [
    "http://localhost:5174/",
    "http://127.0.0.1:5174/",
    "https://admin.photo-gallery.i-dmytro.org/",
  ]
}

variable "billing_alert_email" {
  description = "Email address for AWS Budget and Cost Anomaly Detection alerts. Keep the value in ignored terraform.tfvars."
  type        = string
  sensitive   = true
}

variable "service_anomaly_monitor_arn" {
  description = "Existing account-wide Cost Anomaly Detection SERVICE monitor ARN. Keep the account-specific value in ignored terraform.tfvars."
  type        = string
}
