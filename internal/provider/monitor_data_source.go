// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/norainacloud/terraform-provider-norainadns/internal/client"
)

// Ensure MonitorDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &MonitorDataSource{}
	_ datasource.DataSourceWithConfigure = &MonitorDataSource{}
)

// MonitorDataSource defines the monitor data source implementation
type MonitorDataSource struct {
	client *client.Client
}

// MonitorDataSourceModel describes the monitor data source data model
type MonitorDataSourceModel struct {
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

// NewMonitorDataSource creates a new monitor data source
func NewMonitorDataSource() datasource.DataSource {
	return &MonitorDataSource{}
}

// Metadata returns the data source type name
func (d *MonitorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

// Schema defines the schema for the monitor data source
func (d *MonitorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches information about a health monitor.",
		MarkdownDescription: "Fetches information about a health monitor in NorainaDNS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this monitor.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of this monitor.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of this monitor.",
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol used for health checks.",
				Computed:    true,
			},
			"target_host": schema.StringAttribute{
				Description: "The hostname being monitored.",
				Computed:    true,
			},
			"target_port": schema.Int64Attribute{
				Description: "The port being checked.",
				Computed:    true,
			},
			"target_path": schema.StringAttribute{
				Description: "The path being checked.",
				Computed:    true,
			},
			"check_interval_seconds": schema.Int64Attribute{
				Description: "How often the endpoint is checked.",
				Computed:    true,
			},
			"failure_threshold": schema.Int64Attribute{
				Description: "Number of failures before unhealthy.",
				Computed:    true,
			},
			"timeout_seconds": schema.Int64Attribute{
				Description: "Timeout for each check.",
				Computed:    true,
			},
			"custom_host_header": schema.StringAttribute{
				Description: "Custom Host header.",
				Computed:    true,
			},
			"custom_sni": schema.StringAttribute{
				Description: "Custom SNI.",
				Computed:    true,
			},
			"force_ip_address": schema.StringAttribute{
				Description: "Forced IP address.",
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this monitor is active.",
				Computed:    true,
			},
			"current_status": schema.StringAttribute{
				Description: "The current health status.",
				Computed:    true,
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *MonitorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read reads the monitor data source
func (d *MonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ID.ValueString()

	tflog.Debug(ctx, "Reading monitor data source", map[string]interface{}{
		"id": id,
	})

	monitor, err := d.client.GetMonitor(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Monitor",
			fmt.Sprintf("Could not read monitor ID %s: %s", id, err.Error()),
		)
		return
	}

	// Map response to state
	d.mapMonitorToModel(monitor, &config)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// mapMonitorToModel maps a Monitor API response to the Terraform model
func (d *MonitorDataSource) mapMonitorToModel(monitor *client.Monitor, model *MonitorDataSourceModel) {
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

// =============================================================================
// Monitors Data Source (plural - list all monitors)
// =============================================================================

// Ensure MonitorsDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &MonitorsDataSource{}
	_ datasource.DataSourceWithConfigure = &MonitorsDataSource{}
)

// MonitorsDataSource defines the monitors data source implementation
type MonitorsDataSource struct {
	client *client.Client
}

// MonitorsDataSourceModel describes the monitors data source data model
type MonitorsDataSourceModel struct {
	Monitors []MonitorDataSourceModel `tfsdk:"monitors"`
}

// NewMonitorsDataSource creates a new monitors data source
func NewMonitorsDataSource() datasource.DataSource {
	return &MonitorsDataSource{}
}

// Metadata returns the data source type name
func (d *MonitorsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

// Schema defines the schema for the monitors data source
func (d *MonitorsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all health monitors.",
		MarkdownDescription: "Fetches all health monitors in NorainaDNS.",
		Attributes: map[string]schema.Attribute{
			"monitors": schema.ListNestedAttribute{
				Description: "List of monitors.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description.",
							Computed:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "The protocol.",
							Computed:    true,
						},
						"target_host": schema.StringAttribute{
							Description: "The target host.",
							Computed:    true,
						},
						"target_port": schema.Int64Attribute{
							Description: "The target port.",
							Computed:    true,
						},
						"target_path": schema.StringAttribute{
							Description: "The target path.",
							Computed:    true,
						},
						"check_interval_seconds": schema.Int64Attribute{
							Description: "Check interval.",
							Computed:    true,
						},
						"failure_threshold": schema.Int64Attribute{
							Description: "Failure threshold.",
							Computed:    true,
						},
						"timeout_seconds": schema.Int64Attribute{
							Description: "Timeout.",
							Computed:    true,
						},
						"custom_host_header": schema.StringAttribute{
							Description: "Custom Host header.",
							Computed:    true,
						},
						"custom_sni": schema.StringAttribute{
							Description: "Custom SNI.",
							Computed:    true,
						},
						"force_ip_address": schema.StringAttribute{
							Description: "Forced IP.",
							Computed:    true,
						},
						"is_active": schema.BoolAttribute{
							Description: "Is active.",
							Computed:    true,
						},
						"current_status": schema.StringAttribute{
							Description: "Current status.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *MonitorsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read reads the monitors data source
func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading monitors data source")

	monitors, err := d.client.GetMonitors(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Monitors",
			fmt.Sprintf("Could not read monitors: %s", err.Error()),
		)
		return
	}

	var state MonitorsDataSourceModel
	for _, monitor := range monitors {
		var monitorModel MonitorDataSourceModel
		d.mapMonitorToModelPtr(&monitor, &monitorModel)
		state.Monitors = append(state.Monitors, monitorModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// mapMonitorToModelPtr maps a Monitor API response to the Terraform model (pointer version)
func (d *MonitorsDataSource) mapMonitorToModelPtr(monitor *client.Monitor, model *MonitorDataSourceModel) {
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
