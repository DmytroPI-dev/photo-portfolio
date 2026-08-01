# Cognito owns administrator identities; public gallery visitors never interact
# with this pool. There is no self-service sign-up, so the gallery owner creates
# the small administrator set deliberately through the AWS CLI or console.
resource "aws_cognito_user_pool" "gallery_admin" {
  name                     = local.admin_user_pool
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]
  mfa_configuration        = "OPTIONAL"
  deletion_protection      = "ACTIVE"

  admin_create_user_config {
    allow_admin_create_user_only = true
  }

  password_policy {
    minimum_length                   = 14
    require_lowercase                = true
    require_numbers                  = true
    require_symbols                  = true
    require_uppercase                = true
    temporary_password_validity_days = 7
  }

  software_token_mfa_configuration {
    enabled = true
  }

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }
}

resource "aws_cognito_resource_server" "gallery_api" {
  identifier   = "gallery"
  name         = "Gallery administration API"
  user_pool_id = aws_cognito_user_pool.gallery_admin.id

  scope {
    scope_name        = "read"
    scope_description = "Read administrative gallery metadata."
  }

  scope {
    scope_name        = "write"
    scope_description = "Create and update gallery metadata."
  }

  scope {
    scope_name        = "publish"
    scope_description = "Publish, archive, and reorder gallery work."
  }
}

resource "aws_cognito_user_pool_domain" "gallery_admin" {
  domain       = local.admin_cognito_domain
  user_pool_id = aws_cognito_user_pool.gallery_admin.id
}

# A browser SPA must never receive a client secret. Cognito's authorization-code
# flow with PKCE protects the code exchange, and API Gateway validates the
# resulting access token before any future /admin route reaches the Lambda.
resource "aws_cognito_user_pool_client" "gallery_admin" {
  name                                 = "${local.name_prefix}-gallery-admin-spa"
  user_pool_id                         = aws_cognito_user_pool.gallery_admin.id
  generate_secret                      = false
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes = [
    "openid",
    "email",
    "profile",
    "gallery/read",
    "gallery/write",
    "gallery/publish",
  ]
  callback_urls                 = var.admin_callback_urls
  logout_urls                   = var.admin_logout_urls
  supported_identity_providers  = ["COGNITO"]
  prevent_user_existence_errors = "ENABLED"
  enable_token_revocation       = true
  access_token_validity         = 60
  id_token_validity             = 60
  refresh_token_validity        = 1

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  depends_on = [aws_cognito_resource_server.gallery_api]
}

# It is intentionally not attached to a route yet. The public API stays
# anonymous, while the forthcoming /admin routes will opt into this authorizer
# and the least-privilege scope appropriate to each operation.
resource "aws_apigatewayv2_authorizer" "gallery_admin" {
  api_id           = aws_apigatewayv2_api.gallery.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "${local.name_prefix}-gallery-admin"

  jwt_configuration {
    audience = [aws_cognito_user_pool_client.gallery_admin.id]
    issuer   = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.gallery_admin.id}"
  }
}
