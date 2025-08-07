// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/flpajany/terraform-provider-omni/omniapi"
)

// Ensure OmniProvider satisfies various provider interfaces.
var _ provider.Provider = &OmniProvider{}
var _ provider.ProviderWithFunctions = &OmniProvider{}

// OmniProvider defines the provider implementation.
type OmniProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OmniProviderModel describes the provider data model.
type OmniProviderModel struct {
	Uri            types.String `tfsdk:"uri"`
	ServiceAccount types.String `tfsdk:"service_account"`
}

func (p *OmniProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "omni"
	resp.Version = p.version
}

func (p *OmniProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				MarkdownDescription: "Omni provider attribute",
				Required:            true,
			},
			"service_account": schema.StringAttribute{
				MarkdownDescription: "Omni provider attribute",
				Required:            true,
			},
		},
	}
}

func (p *OmniProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OmniProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := omniapi.NewClient(data.Uri.ValueString(), data.ServiceAccount.ValueString())
	err := client.Open()
	if err != nil {
		resp.Diagnostics.AddError("clien.Open() not working", err.Error())
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OmniProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOmniClusterResource,
		NewOmniKubeconfigResource,
	}
}

func (p *OmniProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOmniMachineDataSource,
		NewOmniTalosconfigDataSource,
	}
}

func (p *OmniProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OmniProvider{
			version: version,
		}
	}
}
