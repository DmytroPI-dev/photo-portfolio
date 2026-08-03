output "api_base_url" {
  description = "Generated API Gateway base URL for the public gallery metadata API."
  value       = aws_apigatewayv2_api.gallery.api_endpoint
}

output "api_function_name" {
  description = "Gallery metadata Lambda function name."
  value       = aws_lambda_function.gallery_api.function_name
}

output "metadata_table_name" {
  description = "DynamoDB table reserved for gallery metadata."
  value       = aws_dynamodb_table.gallery_metadata.name
}

output "originals_bucket_name" {
  description = "Private S3 bucket receiving administrator-uploaded original images."
  value       = aws_s3_bucket.gallery_originals.bucket
}

output "metadata_table_arn" {
  description = "DynamoDB table ARN for the upcoming repository and seed command."
  value       = aws_dynamodb_table.gallery_metadata.arn
}

output "cost_anomaly_monitor_arn" {
  description = "External account-wide Cost Anomaly Detection service monitor ARN used by this project's alert subscription."
  value       = var.service_anomaly_monitor_arn
}

output "admin_cognito_issuer" {
  description = "Cognito issuer used by API Gateway's future protected admin routes."
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.gallery_admin.id}"
}

output "admin_cognito_domain" {
  description = "Cognito hosted-UI base domain for the administrator SPA."
  value       = "https://${aws_cognito_user_pool_domain.gallery_admin.domain}.auth.${var.aws_region}.amazoncognito.com"
}

output "admin_cognito_client_id" {
  description = "Public OAuth client ID for the React-admin SPA; it is not a secret."
  value       = aws_cognito_user_pool_client.gallery_admin.id
}
