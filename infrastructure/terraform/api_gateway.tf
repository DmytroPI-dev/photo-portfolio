resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${local.api_name}"
  retention_in_days = var.log_retention_days
}

resource "aws_apigatewayv2_api" "gallery" {
  name          = local.api_name
  protocol_type = "HTTP"

  cors_configuration {
    allow_credentials = false
    allow_headers     = ["authorization", "content-type"]
    allow_methods     = ["GET", "POST", "PATCH", "DELETE", "OPTIONS"]
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

# The public routes above remain anonymous. Admin reads opt into Cognito and
# the narrow read scope before reaching the shared Lambda integration.
resource "aws_apigatewayv2_route" "admin_collections" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "GET /admin/collections"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/read"]
}

resource "aws_apigatewayv2_route" "admin_collection" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "GET /admin/collections/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/read"]
}

resource "aws_apigatewayv2_route" "admin_collections_write" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/collections"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_collection_write" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "PATCH /admin/collections/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_collection_publish" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/collections/{id}/publish"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

resource "aws_apigatewayv2_route" "admin_collection_archive" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/collections/{id}/archive"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

# Restoring returns an archived collection to a private draft. It cannot make a
# collection public; publication remains a separate, deliberate action.
resource "aws_apigatewayv2_route" "admin_collection_restore" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/collections/{id}/restore"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

# Permanent deletion is intentionally a publish-level lifecycle action. The Go
# handler only permits an archived, empty collection after an exact-slug check.
resource "aws_apigatewayv2_route" "admin_collection_delete" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "DELETE /admin/collections/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

resource "aws_apigatewayv2_route" "admin_photos" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "GET /admin/photos"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/read"]
}

resource "aws_apigatewayv2_route" "admin_uploads" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/uploads"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_photo" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "GET /admin/photos/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/read"]
}

resource "aws_apigatewayv2_route" "admin_photo_preview" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "GET /admin/photos/{id}/preview"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/read"]
}

resource "aws_apigatewayv2_route" "admin_photos_write" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_photo_write" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "PATCH /admin/photos/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_photos_reorder" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos/reorder"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
}

resource "aws_apigatewayv2_route" "admin_photo_publish" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos/{id}/publish"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

resource "aws_apigatewayv2_route" "admin_photo_archive" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos/{id}/archive"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

resource "aws_apigatewayv2_route" "admin_photo_restore" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos/{id}/restore"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

# Permanent deletion is a publish-level lifecycle action. The handler accepts
# it only for archived photos with a version and exact typed ID confirmation.
resource "aws_apigatewayv2_route" "admin_photo_delete" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "DELETE /admin/photos/{id}"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/publish"]
}

# Retrying an existing private original starts no publication transition. It is
# a write action because it schedules worker capacity and changes no public data.
resource "aws_apigatewayv2_route" "admin_photo_retry" {
  api_id               = aws_apigatewayv2_api.gallery.id
  route_key            = "POST /admin/photos/{id}/retry"
  target               = "integrations/${aws_apigatewayv2_integration.gallery_api.id}"
  authorization_type   = "JWT"
  authorizer_id        = aws_apigatewayv2_authorizer.gallery_admin.id
  authorization_scopes = ["gallery/write"]
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
