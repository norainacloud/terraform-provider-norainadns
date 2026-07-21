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

// Ensure ZoneDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &ZoneDataSource{}
	_ datasource.DataSourceWithConfigure = &ZoneDataSource{}
)

// ZoneDataSource defines the zone data source implementation
type ZoneDataSource struct {
	client *client.Client
}

// ZoneDataSourceModel describes the zone data source data model
type ZoneDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	DomainName  types.String `tfsdk:"domain_name"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	SyncStatus  types.String `tfsdk:"sync_status"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

// NewZoneDataSource creates a new zone data source
func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

// Metadata returns the data source type name
func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

// Schema defines the schema for the zone data source
func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches information about a DNS zone.",
		MarkdownDescription: "Fetches information about a DNS zone in NorainaDNS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this zone.",
				Required:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name for this zone.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of this zone.",
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this zone is active.",
				Computed:    true,
			},
			"sync_status": schema.StringAttribute{
				Description: "The synchronization status of this zone.",
				Computed:    true,
			},
			"nameservers": schema.ListAttribute{
				Description: "The list of nameservers for this zone.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read reads the zone data source
func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ZoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ID.ValueString()

	tflog.Debug(ctx, "Reading zone data source", map[string]interface{}{
		"id": id,
	})

	zone, err := d.client.GetZone(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zone",
			fmt.Sprintf("Could not read zone ID %s: %s", id, err.Error()),
		)
		return
	}

	// Get nameservers
	ns, err := d.client.GetZoneNameservers(ctx, zone.ID)
	if err != nil {
		tflog.Warn(ctx, "Could not get nameservers", map[string]interface{}{
			"zone_id": zone.ID,
			"error":   err.Error(),
		})
	}

	// Map response to state
	config.ID = types.StringValue(zone.ID)
	config.DomainName = types.StringValue(zone.DomainName)
	config.IsActive = types.BoolValue(zone.IsActive)
	config.SyncStatus = types.StringValue(zone.SyncStatus)

	if zone.Description != nil {
		config.Description = types.StringValue(*zone.Description)
	} else {
		config.Description = types.StringNull()
	}

	// Set nameservers
	if ns != nil && len(ns.Nameservers) > 0 {
		nameservers, diags := types.ListValueFrom(ctx, types.StringType, ns.Nameservers)
		resp.Diagnostics.Append(diags...)
		config.Nameservers = nameservers
	} else {
		config.Nameservers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// =============================================================================
// Zones Data Source (plural - list all zones)
// =============================================================================

// Ensure ZonesDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &ZonesDataSource{}
	_ datasource.DataSourceWithConfigure = &ZonesDataSource{}
)

// ZonesDataSource defines the zones data source implementation
type ZonesDataSource struct {
	client *client.Client
}

// ZonesDataSourceModel describes the zones data source data model
type ZonesDataSourceModel struct {
	Zones []ZoneDataSourceModel `tfsdk:"zones"`
}

// NewZonesDataSource creates a new zones data source
func NewZonesDataSource() datasource.DataSource {
	return &ZonesDataSource{}
}

// Metadata returns the data source type name
func (d *ZonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zones"
}

// Schema defines the schema for the zones data source
func (d *ZonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all DNS zones.",
		MarkdownDescription: "Fetches all DNS zones in NorainaDNS that the authenticated user has access to.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Description: "List of zones.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for this zone.",
							Computed:    true,
						},
						"domain_name": schema.StringAttribute{
							Description: "The domain name for this zone.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of this zone.",
							Computed:    true,
						},
						"is_active": schema.BoolAttribute{
							Description: "Whether this zone is active.",
							Computed:    true,
						},
						"sync_status": schema.StringAttribute{
							Description: "The synchronization status.",
							Computed:    true,
						},
						"nameservers": schema.ListAttribute{
							Description: "The list of nameservers.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *ZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read reads the zones data source
func (d *ZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading zones data source")

	zones, err := d.client.GetZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zones",
			fmt.Sprintf("Could not read zones: %s", err.Error()),
		)
		return
	}

	var state ZonesDataSourceModel
	for _, zone := range zones {
		zoneModel := ZoneDataSourceModel{
			ID:         types.StringValue(zone.ID),
			DomainName: types.StringValue(zone.DomainName),
			IsActive:   types.BoolValue(zone.IsActive),
			SyncStatus: types.StringValue(zone.SyncStatus),
		}

		if zone.Description != nil {
			zoneModel.Description = types.StringValue(*zone.Description)
		} else {
			zoneModel.Description = types.StringNull()
		}

		// Get nameservers for each zone
		ns, err := d.client.GetZoneNameservers(ctx, zone.ID)
		if err == nil && ns != nil && len(ns.Nameservers) > 0 {
			nameservers, diags := types.ListValueFrom(ctx, types.StringType, ns.Nameservers)
			resp.Diagnostics.Append(diags...)
			zoneModel.Nameservers = nameservers
		} else {
			zoneModel.Nameservers = types.ListNull(types.StringType)
		}

		state.Zones = append(state.Zones, zoneModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
