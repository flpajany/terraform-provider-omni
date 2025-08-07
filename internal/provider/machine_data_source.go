// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/flpajany/terraform-provider-omni/omniapi"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OmniMachineDataSource{}
var _ datasource.DataSourceWithValidateConfig = &OmniMachineDataSource{}

func NewOmniMachineDataSource() datasource.DataSource {
	return &OmniMachineDataSource{}
}

// OmniMachineDataSource defines the data source implementation.
type OmniMachineDataSource struct {
	client *omniapi.OmniClient
}

// OmniMachineDataSourceModel describes the data source data model.
type OmniMachineDataSourceModel struct {
	HardwareAddress     types.String `tfsdk:"hardware_address"`
	UUID                types.String `tfsdk:"uuid"`
	WaitForRegistration types.Bool   `tfsdk:"wait_for_registration"`
	ID                  types.String `tfsdk:"id"`
}

func (d *OmniMachineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (d *OmniMachineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "omni_machine data source",

		Attributes: map[string]schema.Attribute{
			"hardware_address": schema.StringAttribute{
				MarkdownDescription: "hardware_address attribute",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "id identifier",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "uuid identifier",
				Computed:            true,
				Optional:            true,
			},
			"wait_for_registration": schema.BoolAttribute{
				MarkdownDescription: "wait_for_registration flag",
				Optional:            true,
			},
		},
	}
}

func (d *OmniMachineDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OmniMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OmniMachineDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// save into the Terraform state.
	for {
		uuid := data.UUID.ValueString()
		if !data.UUID.IsNull() {
			if machine, ok := d.client.FindMachineByUuid(uuid); ok {
				data.UUID = types.StringValue(machine.Metadata().ID())
				data.ID = types.StringValue(machine.Metadata().ID())
				data.HardwareAddress = types.StringValue(machine.TypedSpec().Value.Network.NetworkLinks[0].HardwareAddress)
				break

			}
		}
		mac := data.HardwareAddress.ValueString()
		if !data.HardwareAddress.IsNull() {
			if machine, ok := d.client.FindMachineByHardwareAddress(mac); ok {
				data.UUID = types.StringValue(machine.Metadata().ID())
				data.ID = types.StringValue(machine.Metadata().ID())
				data.HardwareAddress = types.StringValue(machine.TypedSpec().Value.Network.NetworkLinks[0].HardwareAddress)
				break
			}
		}
		if !data.WaitForRegistration.ValueBool() {
			break
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source omni_machine")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *OmniMachineDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data OmniMachineDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Either uuid or hardware_address should be set
	if data.UUID.IsNull() && data.HardwareAddress.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("hardware_address"), "Missing Attribute Configuration", "hardware_address must be set if uuid is not.")
		return
	}
}
