// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolClientSessionConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientSessionConfigurationConfig_basic(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "sso_session.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sso_session.0.max_validity_minutes", "30"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserPoolClientSessionConfigurationImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrClientID,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClientSessionConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientSessionConfigurationConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "sso_session.0.max_validity_minutes", "0"),
				),
			},
			{
				Config: testAccUserPoolClientSessionConfigurationConfig_basic(rName, 129600),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientSessionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "sso_session.0.max_validity_minutes", "129600"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserPoolClientSessionConfigurationImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrClientID,
			},
		},
	})
}

// The API has no delete operation, so destroying the resource restores the default.
// Terraform must then plan an in-place update rather than a create.
func TestAccCognitoIDPUserPoolClientSessionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientSessionConfigurationConfig_basic(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolClientSessionConfigurationExists(ctx, t, resourceName, &v),
					// nosemgrep:ci.semgrep.acctest.disappears-expect-resource-action -- Session configuration is a sub-resource of the app client; "disappearing" it resets the client to the default configuration, so the plan is an update, not a create.
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPoolClientSessionConfiguration, resourceName),
					testAccCheckUserPoolClientSessionConfigurationIsDefault(ctx, t, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClientSessionConfiguration_disappears_userPoolClient(t *testing.T) {
	ctx := acctest.Context(t)
	var v cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client_session_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientSessionConfigurationConfig_basic(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolClientSessionConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPoolClient, "aws_cognito_user_pool_client.test"),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClientSessionConfiguration_liteTier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientSessionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPoolClientSessionConfigurationConfig_tier(rName, "LITE"),
				ExpectError: regexache.MustCompile(`FeatureUnavailableInTierException`),
			},
		},
	})
}

func testAccCheckUserPoolClientSessionConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_client_session_configuration" {
				continue
			}

			clientID := rs.Primary.Attributes[names.AttrClientID]
			out, err := tfcognitoidp.FindUserPoolClientSessionConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], clientID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, "User Pool Client Session Configuration", clientID, err)
			}

			if !isDefaultUserPoolClientSessionConfiguration(out) {
				return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, "User Pool Client Session Configuration", clientID, errors.New("not reset to defaults"))
			}
		}

		return nil
	}
}

func testAccCheckUserPoolClientSessionConfigurationExists(ctx context.Context, t *testing.T, n string, v *cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolClientSessionConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserPoolClientSessionConfigurationIsDefault(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolClientSessionConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])
		if err != nil {
			return err
		}

		if !isDefaultUserPoolClientSessionConfiguration(output) {
			return fmt.Errorf("User Pool Client Session Configuration (%s) not reset to defaults: max validity minutes %d",
				rs.Primary.Attributes[names.AttrClientID], aws.ToInt32(output.SSOSession.MaxValidityMinutes))
		}

		return nil
	}
}

func isDefaultUserPoolClientSessionConfiguration(v *cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput) bool {
	return aws.ToInt32(v.SSOSession.MaxValidityMinutes) == 60
}

func testAccUserPoolClientSessionConfigurationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrUserPoolID, names.AttrClientID)
}

func testAccUserPoolClientSessionConfigurationConfig_basic(rName string, maxValidityMinutes int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name           = %[1]q
  user_pool_tier = "ESSENTIALS"
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client_session_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  sso_session {
    max_validity_minutes = %[2]d
  }
}
`, rName, maxValidityMinutes)
}

func testAccUserPoolClientSessionConfigurationConfig_tier(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name           = %[1]q
  user_pool_tier = %[2]q
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client_session_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  sso_session {
    max_validity_minutes = 30
  }
}
`, rName, tier)
}
