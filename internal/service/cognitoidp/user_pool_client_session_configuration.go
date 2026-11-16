// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Value that GetUserPoolClientSessionConfiguration returns for an app client that has never been configured.
// Delete restores it because the API has no operation that removes the configuration.
const defaultSSOSessionMaxValidityMinutes = 60

// @FrameworkResource("aws_cognito_user_pool_client_session_configuration", name="User Pool Client Session Configuration")
// @IdentityAttribute("user_pool_id")
// @IdentityAttribute("client_id")
// @ImportIDHandler("userPoolClientSessionConfigurationImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider;cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput")
// @Testing(importStateIdFunc="testAccUserPoolClientSessionConfigurationImportStateIDFunc")
// @Testing(importStateIdAttribute="client_id")
// @Testing(hasNoPreExistingResource=true)
func newUserPoolClientSessionConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &userPoolClientSessionConfigurationResource{}, nil
}

type userPoolClientSessionConfigurationResource struct {
	framework.ResourceWithModel[userPoolClientSessionConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *userPoolClientSessionConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClientID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"sso_session": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ssoSessionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_validity_minutes": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(0, 129600),
							},
						},
					},
				},
			},
		},
	}
}

func (r *userPoolClientSessionConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan userPoolClientSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input cognitoidentityprovider.SetUserPoolClientSessionConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.SetUserPoolClientSessionConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ClientID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userPoolClientSessionConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolClientSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPoolClientSessionConfigurationByTwoPartKey(ctx, conn, state.UserPoolID.ValueString(), state.ClientID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ClientID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userPoolClientSessionConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan, state userPoolClientSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// SetUserPoolClientSessionConfiguration replaces the whole configuration, so send every argument.
		var input cognitoidentityprovider.SetUserPoolClientSessionConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.SetUserPoolClientSessionConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ClientID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *userPoolClientSessionConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolClientSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cognitoidentityprovider.SetUserPoolClientSessionConfigurationInput{
		ClientId:   state.ClientID.ValueStringPointer(),
		UserPoolId: state.UserPoolID.ValueStringPointer(),
		SSOSession: &awstypes.ClientSSOSessionConfigType{
			MaxValidityMinutes: aws.Int32(defaultSSOSessionMaxValidityMinutes),
		},
	}

	_, err := conn.SetUserPoolClientSessionConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ClientID.String())
		return
	}
}

func findUserPoolClientSessionConfigurationByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, clientID string) (*cognitoidentityprovider.GetUserPoolClientSessionConfigurationOutput, error) {
	input := cognitoidentityprovider.GetUserPoolClientSessionConfigurationInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	out, err := conn.GetUserPoolClientSessionConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.SSOSession == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type userPoolClientSessionConfigurationResourceModel struct {
	framework.WithRegionModel
	ClientID   types.String                                     `tfsdk:"client_id"`
	SSOSession fwtypes.ListNestedObjectValueOf[ssoSessionModel] `tfsdk:"sso_session"`
	UserPoolID types.String                                     `tfsdk:"user_pool_id"`
}

type ssoSessionModel struct {
	MaxValidityMinutes types.Int32 `tfsdk:"max_validity_minutes"`
}

var (
	_ inttypes.ImportIDParser = userPoolClientSessionConfigurationImportID{}
)

type userPoolClientSessionConfigurationImportID struct{}

func (userPoolClientSessionConfigurationImportID) Parse(id string) (string, map[string]any, error) {
	userPoolID, clientID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || userPoolID == "" || clientID == "" {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <user-pool-id>"+intflex.ResourceIdSeparator+"<client-id>", id)
	}

	result := map[string]any{
		names.AttrUserPoolID: userPoolID,
		names.AttrClientID:   clientID,
	}

	return id, result, nil
}
