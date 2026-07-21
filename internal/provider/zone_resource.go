// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/norainacloud/terraform-provider-norainadns/internal/client"
)

// Ensure ZoneResource satisfies the resource.Resource interface
var (
	_ resource.Resource                = &ZoneResource{}
	_ resource.ResourceWithConfigure   = &ZoneResource{}
	_ resource.ResourceWithImportState = &ZoneResource{}
)

// ZoneResource defines the zone resource implementation
type ZoneResource struct {
	client *client.Client
}

// ZoneResourceModel describes the zone resource data model
type ZoneResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DomainName  types.String `tfsdk:"domain_name"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	SyncStatus  types.String `tfsdk:"sync_status"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

// NewZoneResource creates a new zone resource
func NewZoneResource() resource.Resource {
	return &ZoneResource{}
}

// Metadata returns the resource type name
func (r *ZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

// Schema defines the schema for the zone resource
func (r *ZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a DNS zone in NorainaDNS.",
		MarkdownDescription: "Manages a DNS zone in NorainaDNS. Zones are automatically synchronized to both AWS Route53 and GCP Cloud DNS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this zone.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_name": schema.StringAttribute{
				Description:         "The domain name for this zone (e.g., 'example.com').",
				MarkdownDescription: "The domain name for this zone (e.g., `example.com`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of this zone.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this zone is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"sync_status": schema.StringAttribute{
				Description: "The synchronization status of this zone (pending, synced, error, partial).",
				Computed:    true,
			},
			"nameservers": schema.ListAttribute{
				Description:         "The list of nameservers for this zone. Use these to configure your domain registrar.",
				MarkdownDescription: "The list of nameservers for this zone. Use these to configure your domain registrar.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure configures the resource with the provider client
func (r *ZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new zone
func (r *ZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating zone", map[string]interface{}{
		"domain_name": plan.DomainName.ValueString(),
	})

	// Create the zone via API
	createReq := client.CreateZoneRequest{
		DomainName:  plan.DomainName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	zone, err := r.client.CreateZone(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Zone",
			fmt.Sprintf("Could not create zone %s: %s", plan.DomainName.ValueString(), err.Error()),
		)
		return
	}

	// Get nameservers
	ns, err := r.client.GetZoneNameservers(ctx, zone.ID)
	if err != nil {
		tflog.Warn(ctx, "Could not get nameservers", map[string]interface{}{
			"zone_id": zone.ID,
			"error":   err.Error(),
		})
	}

	// Map response to state
	plan.ID = types.StringValue(zone.ID)
	plan.DomainName = types.StringValue(zone.DomainName)
	plan.IsActive = types.BoolValue(zone.IsActive)
	plan.SyncStatus = types.StringValue(zone.SyncStatus)

	if zone.Description != nil {
		plan.Description = types.StringValue(*zone.Description)
	}

	// Set nameservers
	if ns != nil && len(ns.Nameservers) > 0 {
		nameservers, diags := types.ListValueFrom(ctx, types.StringType, ns.Nameservers)
		resp.Diagnostics.Append(diags...)
		plan.Nameservers = nameservers
	} else {
		plan.Nameservers = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Created zone", map[string]interface{}{
		"id": zone.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the current state of a zone
func (r *ZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading zone", map[string]interface{}{
		"id": id,
	})

	zone, err := r.client.GetZone(ctx, id)
	if err != nil {
		// If the zone doesn't exist, remove it from state
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "ZONE_NOT_FOUND" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Zone",
			fmt.Sprintf("Could not read zone ID %s: %s", id, err.Error()),
		)
		return
	}

	// Get nameservers
	ns, err := r.client.GetZoneNameservers(ctx, zone.ID)
	if err != nil {
		tflog.Warn(ctx, "Could not get nameservers", map[string]interface{}{
			"zone_id": zone.ID,
			"error":   err.Error(),
		})
	}

	// Map response to state
	state.ID = types.StringValue(zone.ID)
	state.DomainName = types.StringValue(zone.DomainName)
	state.IsActive = types.BoolValue(zone.IsActive)
	state.SyncStatus = types.StringValue(zone.SyncStatus)

	if zone.Description != nil {
		state.Description = types.StringValue(*zone.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Set nameservers
	if ns != nil && len(ns.Nameservers) > 0 {
		nameservers, diags := types.ListValueFrom(ctx, types.StringType, ns.Nameservers)
		resp.Diagnostics.Append(diags...)
		state.Nameservers = nameservers
	} else {
		state.Nameservers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates an existing zone
func (r *ZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	tflog.Debug(ctx, "Updating zone", map[string]interface{}{
		"id": id,
	})

	// Update the zone via API
	isActive := plan.IsActive.ValueBool()
	updateReq := client.UpdateZoneRequest{
		Description: plan.Description.ValueString(),
		IsActive:    &isActive,
	}

	zone, err := r.client.UpdateZone(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zone",
			fmt.Sprintf("Could not update zone ID %s: %s", id, err.Error()),
		)
		return
	}

	// Get nameservers
	ns, err := r.client.GetZoneNameservers(ctx, zone.ID)
	if err != nil {
		tflog.Warn(ctx, "Could not get nameservers", map[string]interface{}{
			"zone_id": zone.ID,
			"error":   err.Error(),
		})
	}

	// Map response to state
	plan.ID = types.StringValue(zone.ID)
	plan.DomainName = types.StringValue(zone.DomainName)
	plan.IsActive = types.BoolValue(zone.IsActive)
	plan.SyncStatus = types.StringValue(zone.SyncStatus)

	if zone.Description != nil {
		plan.Description = types.StringValue(*zone.Description)
	}

	// Set nameservers
	if ns != nil && len(ns.Nameservers) > 0 {
		nameservers, diags := types.ListValueFrom(ctx, types.StringType, ns.Nameservers)
		resp.Diagnostics.Append(diags...)
		plan.Nameservers = nameservers
	} else {
		plan.Nameservers = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Updated zone", map[string]interface{}{
		"id": zone.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes a zone
func (r *ZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting zone", map[string]interface{}{
		"id": id,
	})

	err := r.client.DeleteZone(ctx, id)
	if err != nil {
		// If the zone is already deleted, that's fine
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "ZONE_NOT_FOUND" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Zone",
			fmt.Sprintf("Could not delete zone ID %s: %s", id, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleted zone", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing zone into Terraform state
func (r *ZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
