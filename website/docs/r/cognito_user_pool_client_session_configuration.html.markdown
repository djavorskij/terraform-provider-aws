---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client_session_configuration"
description: |-
  Manages the SSO session settings for an AWS Cognito User Pool app client.
---

# Resource: aws_cognito_user_pool_client_session_configuration

Manages the SSO session settings for an AWS Cognito User Pool app client: how long the app client silently reuses an existing SSO session before it requires the user to authenticate again. For more information, see [SetUserPoolClientSessionConfiguration](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_SetUserPoolClientSessionConfiguration.html) in the Amazon Cognito API Reference.

~> **NOTE:** The setting takes effect only while session tracking is enabled for the user pool. Use the `aws_cognito_user_pool_session_configuration` resource to enable it.

~> **NOTE:** Session configuration requires a user pool on the Essentials or Plus [feature plan](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-sign-in-feature-plans.html).

~> **NOTE:** Cognito has no API to remove an app client session configuration. Destroying this resource resets `max_validity_minutes` to the default of `60`.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "example" {
  name           = "example"
  user_pool_tier = "ESSENTIALS"
}

resource "aws_cognito_user_pool_client" "example" {
  name         = "example"
  user_pool_id = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool_client_session_configuration" "example" {
  user_pool_id = aws_cognito_user_pool.example.id
  client_id    = aws_cognito_user_pool_client.example.id

  sso_session {
    max_validity_minutes = 30
  }
}
```

## Argument Reference

The following arguments are required:

* `client_id` - (Required) ID of the user pool app client.
* `sso_session` - (Required) SSO session settings for the app client. See [`sso_session`](#sso_session) below.
* `user_pool_id` - (Required) ID of the user pool that contains the app client.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `sso_session`

* `max_validity_minutes` - (Required) Maximum time, in minutes, that the app client silently reuses an existing SSO session, measured from the user's last fresh authentication. Valid values are between `0` and `129600` (90 days). `0` means the app client never silently reuses the SSO session and always re-authenticates the user; other app clients in the same user pool can still reuse it.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cognito_user_pool_client_session_configuration.example
  identity = {
    user_pool_id = "us-west-2_example123"
    client_id    = "1example23456789"
  }
}

resource "aws_cognito_user_pool_client_session_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `client_id` (String) ID of the Cognito User Pool app client.
* `user_pool_id` (String) ID of the Cognito User Pool.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Cognito User Pool Client Session Configuration using the `user_pool_id` and `client_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cognito_user_pool_client_session_configuration.example
  id = "us-west-2_example123,1example23456789"
}
```

Using `terraform import`, import a Cognito User Pool Client Session Configuration using the `user_pool_id` and `client_id` separated by a comma (`,`). For example:

```console
% terraform import aws_cognito_user_pool_client_session_configuration.example us-west-2_example123,1example23456789
```
