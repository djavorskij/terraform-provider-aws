resource "aws_cognito_user_pool_client_session_configuration" "test" {
  {{- template "region" }}
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  sso_session {
    max_validity_minutes = 30
  }
}

resource "aws_cognito_user_pool_client" "test" {
  {{- template "region" }}
  name         = var.rName
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  {{- template "region" }}
  name           = var.rName
  user_pool_tier = "ESSENTIALS"
}
