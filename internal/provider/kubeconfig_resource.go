// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/flpajany/terraform-provider-omni/omniapi"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OmniKubeconfigResource{}
var _ resource.ResourceWithImportState = &OmniKubeconfigResource{}

func NewOmniKubeconfigResource() resource.Resource {
	return &OmniKubeconfigResource{}
}

// OmniKubeconfigResource defines the data source implementation.
type OmniKubeconfigResource struct {
	client *omniapi.OmniClient
}

// OmniKubeconfigDataSourceModel describes the data source data model.
type OmniKubeconfigResourceModel struct {
	Kubeconfig  types.String `tfsdk:"kubeconfig"`
	User        types.String `tfsdk:"user"`
	Groups      types.List   `tfsdk:"groups"`
	ClusterName types.String `tfsdk:"cluster_name"`
	ID          types.String `tfsdk:"id"`
}

func (r *OmniKubeconfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubeconfig"
}

func (r *OmniKubeconfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "omni_kubeconfig resource",

		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "cluster_name attribute",
				Required:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "user attribute for kubeconfig",
				Optional:            true,
			},
			"groups": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "groups attribute to use in kubeconfig",
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "kubeconfig attribute",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "id identifier",
				Computed:            true,
			},
		},
	}
}

func (r *OmniKubeconfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*omniapi.OmniClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *omniapi.OmniClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OmniKubeconfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OmniKubeconfigResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user string
	if data.User.IsNull() {
		user = "admin"
	} else {
		user = data.User.ValueString()
	}

	var groups []string
	if data.Groups.IsNull() {
		groups = []string{"system:masters"}
	} else {
		var l []string
		diag := data.Groups.ElementsAs(ctx, &l, true)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		groups = l
	}

	k, err := r.client.GetKubeconfigWithoutOIDC(data.ClusterName.ValueString(), user, groups...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving kubeconfig", fmt.Sprintf("error : %v", err))
	}

	data.Kubeconfig = types.StringValue(k)
	data.ID = data.ClusterName

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "create a resource omni_kubeconfig")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniKubeconfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OmniKubeconfigResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a resource omni_kubeconfig")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniKubeconfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OmniKubeconfigResourceModel
	var plan OmniKubeconfigResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var user string
	if plan.User.IsNull() {
		user = "admin"
	} else {
		user = plan.User.ValueString()
	}

	var groups []string
	if plan.Groups.IsNull() {
		groups = []string{"system:masters"}
	} else {
		var l []string
		diag := plan.Groups.ElementsAs(ctx, &l, true)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		groups = l
	}

	k, err := r.client.GetKubeconfigWithoutOIDC(data.ClusterName.ValueString(), user, groups...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving kubeconfig", fmt.Sprintf("error : %v", err))
	}

	data.Kubeconfig = types.StringValue(k)
	data.User = plan.User
	data.Groups = plan.Groups
	data.ID = data.ClusterName

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "update a resource omni_kubeconfig")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniKubeconfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OmniKubeconfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "delete a resource omni_kubeconfig")
}

func (r *OmniKubeconfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
