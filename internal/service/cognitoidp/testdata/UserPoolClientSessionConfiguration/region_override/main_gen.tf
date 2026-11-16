# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cognito_user_pool_client_session_configuration" "test" {
  region = var.region

  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  sso_session {
    max_validity_minutes = 30
  }
}

resource "aws_cognito_user_pool_client" "test" {
  region = var.region

  name         = var.rName
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  region = var.region

  name           = var.rName
  user_pool_tier = "ESSENTIALS"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
