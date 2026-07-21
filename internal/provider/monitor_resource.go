// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/norainacloud/terraform-provider-norainadns/internal/client"
)

// Ensure MonitorResource satisfies the resource.Resource interface
var (
	_ resource.Resource                = &MonitorResource{}
	_ resource.ResourceWithConfigure   = &MonitorResource{}
	_ resource.ResourceWithImportState = &MonitorResource{}
)

// MonitorResource defines the monitor resource implementation
type MonitorResource struct {
	client *client.Client
}

// MonitorResourceModel describes the monitor resource data model
type MonitorResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Protocol             types.String `tfsdk:"protocol"`
	TargetHost           types.String `tfsdk:"target_host"`
	TargetPort           types.Int64  `tfsdk:"target_port"`
	TargetPath           types.String `tfsdk:"target_path"`
	CheckIntervalSeconds types.Int64  `tfsdk:"check_interval_seconds"`
	FailureThreshold     types.Int64  `tfsdk:"failure_threshold"`
	TimeoutSeconds       types.Int64  `tfsdk:"timeout_seconds"`
	CustomHostHeader     types.String `tfsdk:"custom_host_header"`
	CustomSNI            types.String `tfsdk:"custom_sni"`
	ForceIPAddress       types.String `tfsdk:"force_ip_address"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	CurrentStatus        types.String `tfsdk:"current_status"`
}

// NewMonitorResource creates a new monitor resource
func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

// Metadata returns the resource type name
func (r *MonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

// Schema defines the schema for the monitor resource
func (r *MonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a health monitor in NorainaDNS.",
		MarkdownDescription: `Manages a health monitor in NorainaDNS.

Monitors are used to check the health of your endpoints. When a monitor detects that an endpoint is unhealthy,
DNS records using failover routing will automatically switch to the backup value.

~> **Important:** Monitors must be created before they can be referenced in DNS records with failover routing.
Similarly, you must update DNS records to remove the monitor reference before deleting the monitor.

## CDN/WAF Bypass Use Case

Use the ` + "`force_ip_address`" + ` and ` + "`custom_host_header`" + ` attributes to monitor your origin server
directly, bypassing CDN or WAF. This is useful for failover scenarios where you want to switch traffic
to the origin when the CDN fails.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A descriptive name for this monitor.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of this monitor.",
				Optional:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol to use for health checks (HTTP, HTTPS, or TCP).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS", "TCP"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_host": schema.StringAttribute{
				Description:         "The hostname to monitor.",
				MarkdownDescription: "The hostname to monitor. This is the actual host that will be checked.",
				Required:            true,
			},
			"target_port": schema.Int64Attribute{
				Description: "The port to check. Defaults to 80 for HTTP, 443 for HTTPS.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"target_path": schema.StringAttribute{
				Description: "The path to check (for HTTP/HTTPS). Defaults to '/'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
			},
			"check_interval_seconds": schema.Int64Attribute{
				Description: "How often to check the endpoint, in seconds. Defaults to 60.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(60, 600),
				},
			},
			"failure_threshold": schema.Int64Attribute{
				Description: "Number of consecutive failures before marking as unhealthy. Defaults to 3.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"timeout_seconds": schema.Int64Attribute{
				Description: "Timeout for each health check request, in seconds. Defaults to 10.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(5, 30),
				},
			},
			"custom_host_header": schema.StringAttribute{
				Description:         "Custom Host header to send with HTTP/HTTPS requests.",
				MarkdownDescription: "Custom Host header to send with HTTP/HTTPS requests. Use this when monitoring an origin server that expects a specific hostname (e.g., when bypassing a CDN).",
				Optional:            true,
			},
			"custom_sni": schema.StringAttribute{
				Description:         "Custom SNI (Server Name Indication) for HTTPS requests.",
				MarkdownDescription: "Custom SNI (Server Name Indication) for HTTPS requests. Use this when the SSL certificate expects a different hostname than `target_host`.",
				Optional:            true,
			},
			"force_ip_address": schema.StringAttribute{
				Description:         "IP address to use instead of DNS resolution.",
				MarkdownDescription: "IP address to use instead of DNS resolution. Use this to monitor a specific server IP directly, bypassing DNS (useful for CDN/WAF bypass scenarios).",
				Optional:            true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this monitor is currently active.",
				Computed:    true,
			},
			"current_status": schema.StringAttribute{
				Description: "The current health status (healthy, unhealthy, unknown).",
				Computed:    true,
			},
		},
	}
}

// Configure configures the resource with the provider client
func (r *MonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new monitor
func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating monitor", map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"target_host": plan.TargetHost.ValueString(),
	})

	// Build the create request
	createReq := client.CreateMonitorRequest{
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueString(),
		Protocol:             plan.Protocol.ValueString(),
		TargetHost:           plan.TargetHost.ValueString(),
		TargetPort:           int(plan.TargetPort.ValueInt64()),
		TargetPath:           plan.TargetPath.ValueString(),
		CheckIntervalSeconds: int(plan.CheckIntervalSeconds.ValueInt64()),
		FailureThreshold:     int(plan.FailureThreshold.ValueInt64()),
		TimeoutSeconds:       int(plan.TimeoutSeconds.ValueInt64()),
		CustomHostHeader:     plan.CustomHostHeader.ValueString(),
		CustomSNI:            plan.CustomSNI.ValueString(),
		ForceIPAddress:       plan.ForceIPAddress.ValueString(),
	}

	monitor, err := r.client.CreateMonitor(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Monitor",
			fmt.Sprintf("Could not create monitor %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map response to state
	r.mapMonitorToModel(monitor, &plan)

	tflog.Debug(ctx, "Created monitor", map[string]interface{}{
		"id": monitor.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the current state of a monitor
func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading monitor", map[string]interface{}{
		"id": id,
	})

	monitor, err := r.client.GetMonitor(ctx, id)
	if err != nil {
		// If the monitor doesn't exist, remove it from state
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "MONITOR_NOT_FOUND" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Monitor",
			fmt.Sprintf("Could not read monitor ID %s: %s", id, err.Error()),
		)
		return
	}

	// Map response to state
	r.mapMonitorToModel(monitor, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates an existing monitor
func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	tflog.Debug(ctx, "Updating monitor", map[string]interface{}{
		"id": id,
	})

	// Build the update request
	updateReq := client.UpdateMonitorRequest{
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueString(),
		TargetHost:           plan.TargetHost.ValueString(),
		TargetPort:           int(plan.TargetPort.ValueInt64()),
		TargetPath:           plan.TargetPath.ValueString(),
		CheckIntervalSeconds: int(plan.CheckIntervalSeconds.ValueInt64()),
		FailureThreshold:     int(plan.FailureThreshold.ValueInt64()),
		TimeoutSeconds:       int(plan.TimeoutSeconds.ValueInt64()),
		CustomHostHeader:     plan.CustomHostHeader.ValueString(),
		CustomSNI:            plan.CustomSNI.ValueString(),
		ForceIPAddress:       plan.ForceIPAddress.ValueString(),
	}

	monitor, err := r.client.UpdateMonitor(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Monitor",
			fmt.Sprintf("Could not update monitor ID %s: %s", id, err.Error()),
		)
		return
	}

	// Map response to state
	r.mapMonitorToModel(monitor, &plan)

	tflog.Debug(ctx, "Updated monitor", map[string]interface{}{
		"id": monitor.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes a monitor
func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting monitor", map[string]interface{}{
		"id": id,
	})

	err := r.client.DeleteMonitor(ctx, id)
	if err != nil {
		// If the monitor is already deleted, that's fine
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "MONITOR_NOT_FOUND" {
			return
		}
		// Check if monitor is still in use by records
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "MONITOR_IN_USE" {
			resp.Diagnostics.AddError(
				"Error Deleting Monitor",
				fmt.Sprintf("Cannot delete monitor ID %s because it is still referenced by DNS records. "+
					"Remove the monitor reference from the records first, then delete the monitor.", id),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Monitor",
			fmt.Sprintf("Could not delete monitor ID %s: %s", id, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleted monitor", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing monitor into Terraform state
func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapMonitorToModel maps a Monitor API response to the Terraform model
func (r *MonitorResource) mapMonitorToModel(monitor *client.Monitor, model *MonitorResourceModel) {
	model.ID = types.StringValue(monitor.ID)
	model.Name = types.StringValue(monitor.Name)
	model.Protocol = types.StringValue(monitor.Protocol)
	model.TargetHost = types.StringValue(monitor.TargetHost)
	model.TargetPort = types.Int64Value(int64(monitor.TargetPort))
	model.TargetPath = types.StringValue(monitor.TargetPath)
	model.CheckIntervalSeconds = types.Int64Value(int64(monitor.CheckIntervalSeconds))
	model.FailureThreshold = types.Int64Value(int64(monitor.FailureThreshold))
	model.TimeoutSeconds = types.Int64Value(int64(monitor.TimeoutSeconds))
	model.IsActive = types.BoolValue(monitor.IsActive)
	model.CurrentStatus = types.StringValue(monitor.CurrentStatus)

	if monitor.Description != nil {
		model.Description = types.StringValue(*monitor.Description)
	} else {
		model.Description = types.StringNull()
	}
	if monitor.CustomHostHeader != nil {
		model.CustomHostHeader = types.StringValue(*monitor.CustomHostHeader)
	} else {
		model.CustomHostHeader = types.StringNull()
	}
	if monitor.CustomSNI != nil {
		model.CustomSNI = types.StringValue(*monitor.CustomSNI)
	} else {
		model.CustomSNI = types.StringNull()
	}
	if monitor.ForceIPAddress != nil {
		model.ForceIPAddress = types.StringValue(*monitor.ForceIPAddress)
	} else {
		model.ForceIPAddress = types.StringNull()
	}
}
