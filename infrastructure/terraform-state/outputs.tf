output "state_bucket_name" {
  description = "S3 bucket name for the gallery Terraform state."
  value       = aws_s3_bucket.terraform_state.bucket
}

output "state_key" {
  description = "Recommended application Terraform state object key."
  value       = "${var.project}/${var.environment}/terraform.tfstate"
}

output "backend_config_hcl" {
  description = "Backend configuration to place in infrastructure/terraform/backend.hcl."
  value       = <<-EOT
    bucket       = "${aws_s3_bucket.terraform_state.bucket}"
    key          = "${var.project}/${var.environment}/terraform.tfstate"
    region        = "${var.aws_region}"
    use_lockfile  = true
    encrypt       = true
  EOT
}
