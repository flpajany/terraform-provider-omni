// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/flpajany/terraform-provider-omni/omniapi"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OmniTalosconfigDataSource{}
var _ datasource.DataSourceWithValidateConfig = &OmniTalosconfigDataSource{}

func NewOmniTalosconfigDataSource() datasource.DataSource {
	return &OmniTalosconfigDataSource{}
}

// OmniTalosconfigDataSource defines the data source implementation.
type OmniTalosconfigDataSource struct {
	client *omniapi.OmniClient
}

// OmniTalosconfigDataSourceModel describes the data source data model.
type OmniTalosconfigDataSourceModel struct {
	Talosconfig types.String `tfsdk:"talosconfig"`
	ClusterName types.String `tfsdk:"cluster_name"`
	ID          types.String `tfsdk:"id"`
}

func (d *OmniTalosconfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_talosconfig"
}

func (d *OmniTalosconfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "omni_talosconfig data source",

		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "cluster_name attribute",
				Required:            true,
			},
			"talosconfig": schema.StringAttribute{
				MarkdownDescription: "talosconfig attribute",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "id identifier",
				Computed:            true,
			},
		},
	}
}

func (d *OmniTalosconfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*omniapi.OmniClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *omniapi.OmniClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OmniTalosconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OmniTalosconfigDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	k, err := d.client.GetTalosconfig(data.ClusterName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving talosconfig", fmt.Sprintf("error : %v", err))
	}

	data.Talosconfig = types.StringValue(k)
	data.ID = data.ClusterName

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source omni_talosconfig")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *OmniTalosconfigDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data OmniTalosconfigDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
