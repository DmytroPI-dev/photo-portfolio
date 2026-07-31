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

output "metadata_table_arn" {
  description = "DynamoDB table ARN for the upcoming repository and seed command."
  value       = aws_dynamodb_table.gallery_metadata.arn
}
