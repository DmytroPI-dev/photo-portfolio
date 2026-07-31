resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${local.api_name}"
  retention_in_days = var.log_retention_days
}

resource "aws_apigatewayv2_api" "gallery" {
  name          = local.api_name
  protocol_type = "HTTP"

  cors_configuration {
    allow_credentials = false
    allow_headers     = ["content-type"]
    allow_methods     = ["GET", "OPTIONS"]
    allow_origins     = var.allowed_cors_origins
    max_age           = 3600
  }
}

resource "aws_apigatewayv2_integration" "gallery_api" {
  api_id = aws_apigatewayv2_api.gallery.id

  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.gallery_api.invoke_arn
  integration_method     = "POST"
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "health" {
  api_id    = aws_apigatewayv2_api.gallery.id
  route_key = "GET /health"
  target    = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
}

resource "aws_apigatewayv2_route" "collections" {
  api_id    = aws_apigatewayv2_api.gallery.id
  route_key = "GET /collections"
  target    = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
}

resource "aws_apigatewayv2_route" "collection" {
  api_id    = aws_apigatewayv2_api.gallery.id
  route_key = "GET /collections/{slug}"
  target    = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
}

resource "aws_apigatewayv2_route" "photo" {
  api_id    = aws_apigatewayv2_api.gallery.id
  route_key = "GET /photos/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.gallery.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      requestTime    = "$context.requestTime"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
      integrationErr = "$context.integrationErrorMessage"
    })
  }
}

resource "aws_lambda_permission" "allow_api_gateway" {
  statement_id  = "AllowExecutionFromApiGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.gallery_api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.gallery.execution_arn}/*/*"
}
