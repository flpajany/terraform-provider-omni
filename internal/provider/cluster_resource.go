// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/flpajany/terraform-provider-omni/omniapi"
	"gopkg.in/yaml.v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OmniClusterResource{}
var _ resource.ResourceWithImportState = &OmniClusterResource{}

func NewOmniClusterResource() resource.Resource {
	return &OmniClusterResource{}
}

// OmniClusterResource defines the resource implementation.
type OmniClusterResource struct {
	client *omniapi.OmniClient
}

// OmniClusterResourceModel describes the resource data model.
type OmniClusterResourceModel struct {
	Template              types.String `tfsdk:"template"`
	TemplateComputed      types.String `tfsdk:"template_computed"`
	ID                    types.String `tfsdk:"id"`
	ForceManifestUpdating types.Bool   `tfsdk:"force_manifest_updating"`
	DeleteMachineLinks    types.Bool   `tfsdk:"delete_machine_links"`
}

func (r *OmniClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *OmniClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "omni_cluster resource",

		Attributes: map[string]schema.Attribute{
			"template": schema.StringAttribute{
				MarkdownDescription: "Template in YAML for managing Omni Cluster",
				Required:            true,
			},
			"template_computed": schema.StringAttribute{
				MarkdownDescription: "Template in YAML formatted by Omni",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID",
				Computed:            true,
			},
			"force_manifest_updating": schema.BoolAttribute{
				MarkdownDescription: "When updating a template, apply automatically updates to manifests",
				Optional:            true,
			},
			"delete_machine_links": schema.BoolAttribute{
				MarkdownDescription: "When destroying a cluster, delete machine links too",
				Optional:            true,
			},
		},
	}
}

func (r *OmniClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*omniapi.OmniClient)

	if !ok {
		// à modifier
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *omniapi.OmniClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OmniClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OmniClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SyncClusterAndWaitForReady(strings.NewReader(data.Template.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to sync cluster, got error: %s", err))
		return
	}

	name, err := r.client.GetClusterNameFromTemplate(strings.NewReader(data.Template.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster name, got error: %s", err))
		return
	}

	template, err := r.client.GetTemplateFromClusterName(name)
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster template, got error: %s", err))
		return
	}

	data.TemplateComputed = types.StringValue(template)
	data.ID = types.StringValue(name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource omni_cluster")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OmniClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	template, err := r.client.GetTemplateFromClusterName(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster template, got error: %s", err))
		return
	}

	data.TemplateComputed = types.StringValue(template)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OmniClusterResourceModel
	var state OmniClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if yes, err := isClusterChangingName(data.Template.ValueString(), state.Template.ValueString()); yes || err != nil {
		if err != nil {
			resp.Diagnostics.AddError("Error parsing template", "Problem with YAML parsing")
			return
		}
		resp.Diagnostics.AddError("Changing Cluster Name is not possible", "Need to destroy resource before create it again")
		return
	}

	err := r.client.SyncCluster(strings.NewReader(data.Template.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to sync cluster, got error: %s", err))
		return
	}

	name, err := r.client.GetClusterNameFromTemplate(strings.NewReader(data.Template.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster name, got error: %s", err))
		return
	}

	if data.ForceManifestUpdating.ValueBool() {
		if err := r.client.SyncManifests(name); err != nil {
			resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to force sync manifests, got error: %s", err))
			return
		}
	}

	template, err := r.client.GetTemplateFromClusterName(name)
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster template, got error: %s", err))
		return
	}

	data.TemplateComputed = types.StringValue(template)
	data.ID = types.StringValue(name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OmniClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OmniClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, err := r.client.GetClusterNameFromTemplate(strings.NewReader(data.Template.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get cluster name, got error: %v", err))
		return
	}

	machinesToDelete, err := r.client.GetClusterMachines(name)
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to get machines associated to cluster %s, got error: %v", name, err))
		return
	}

	err = r.client.DeleteCluster(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to delete cluster, got error: %s", err))
		return
	}

	if data.DeleteMachineLinks.ValueBool() {
		if err := r.client.DeleteClusterMachines(machinesToDelete); err != nil {
			resp.Diagnostics.AddError("client Error", fmt.Sprintf("unable to delete cluster machines links, got error: %s", err))
			return
		}
	}

}

func (r *OmniClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func isClusterChangingName(planTemplate, stateTemplate string) (bool, error) {
	T := struct {
		Kind string
		Name string
	}{}

	var planName, stateName string

	d := yaml.NewDecoder(strings.NewReader(planTemplate))
	for {
		err := d.Decode(&T)
		if err != nil {
			return false, fmt.Errorf("cluster name not found : %v", err)
		}
		if T.Kind == "Cluster" {
			planName = T.Name
			break
		}
	}
	d = yaml.NewDecoder(strings.NewReader(stateTemplate))
	for {
		err := d.Decode(&T)
		if err != nil {
			return false, fmt.Errorf("cluster name not found : %v", err)
		}
		if T.Kind == "Cluster" {
			stateName = T.Name
			break
		}
	}
	if planName != stateName {
		return true, nil
	}
	return false, nil
}
