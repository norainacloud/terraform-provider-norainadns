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

// Ensure RecordDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &RecordDataSource{}
	_ datasource.DataSourceWithConfigure = &RecordDataSource{}
)

// RecordDataSource defines the record data source implementation
type RecordDataSource struct {
	client *client.Client
}

// RecordDataSourceModel describes the record data source data model
type RecordDataSourceModel struct {
	ID            types.String           `tfsdk:"id"`
	ZoneID        types.String           `tfsdk:"zone_id"`
	Name          types.String           `tfsdk:"name"`
	Type          types.String           `tfsdk:"type"`
	TTL           types.Int64            `tfsdk:"ttl"`
	IsDynamic     types.Bool             `tfsdk:"is_dynamic"`
	RoutingPolicy types.String           `tfsdk:"routing_policy"`
	Values        []RecordValueDataModel `tfsdk:"values"`
	SyncStatus    types.String           `tfsdk:"sync_status"`
	FQDN          types.String           `tfsdk:"fqdn"`
}

// RecordValueDataModel describes a record value in data sources
type RecordValueDataModel struct {
	Value     types.String `tfsdk:"value"`
	Priority  types.Int64  `tfsdk:"priority"`
	Weight    types.Int64  `tfsdk:"weight"`
	MonitorID types.String `tfsdk:"monitor_id"`
}

// NewRecordDataSource creates a new record data source
func NewRecordDataSource() datasource.DataSource {
	return &RecordDataSource{}
}

// Metadata returns the data source type name
func (d *RecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

// Schema defines the schema for the record data source
func (d *RecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches information about a DNS record.",
		MarkdownDescription: "Fetches information about a DNS record in NorainaDNS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this record.",
				Required:    true,
			},
			"zone_id": schema.StringAttribute{
				Description: "The ID of the zone this record belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the record.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The record type.",
				Computed:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds.",
				Computed:    true,
			},
			"is_dynamic": schema.BoolAttribute{
				Description: "Whether this is a dynamic record.",
				Computed:    true,
			},
			"routing_policy": schema.StringAttribute{
				Description: "The routing policy.",
				Computed:    true,
			},
			"sync_status": schema.StringAttribute{
				Description: "The synchronization status.",
				Computed:    true,
			},
			"fqdn": schema.StringAttribute{
				Description: "The fully qualified domain name.",
				Computed:    true,
			},
			"values": schema.ListNestedAttribute{
				Description: "The values for this record.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Description: "The record value.",
							Computed:    true,
						},
						"priority": schema.Int64Attribute{
							Description: "The priority.",
							Computed:    true,
						},
						"weight": schema.Int64Attribute{
							Description: "The weight.",
							Computed:    true,
						},
						"monitor_id": schema.StringAttribute{
							Description: "The monitor ID.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *RecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read reads the record data source
func (d *RecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RecordDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID := config.ZoneID.ValueString()

	recordID := config.ID.ValueString()

	tflog.Debug(ctx, "Reading record data source", map[string]interface{}{
		"zone_id":   zoneID,
		"record_id": recordID,
	})

	record, err := d.client.GetRecord(ctx, zoneID, recordID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Record",
			fmt.Sprintf("Could not read record ID %s: %s", recordID, err.Error()),
		)
		return
	}

	// Get zone info for FQDN
	zone, _ := d.client.GetZone(ctx, zoneID)
	fqdn := d.buildFQDN(record.Name, zone)

	// Map response to state
	d.mapRecordToModel(record, &config, fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// buildFQDN builds the fully qualified domain name
func (d *RecordDataSource) buildFQDN(name string, zone *client.Zone) string {
	if zone == nil {
		return ""
	}
	if name == "@" {
		return zone.DomainName + "."
	}
	return name + "." + zone.DomainName + "."
}

// mapRecordToModel maps a Record API response to the data source model
func (d *RecordDataSource) mapRecordToModel(record *client.Record, model *RecordDataSourceModel, fqdn string) {
	model.ID = types.StringValue(record.ID)
	model.ZoneID = types.StringValue(record.ZoneID)
	model.Name = types.StringValue(record.Name)
	model.Type = types.StringValue(record.RecordType)
	model.TTL = types.Int64Value(record.TTL)
	model.IsDynamic = types.BoolValue(record.IsDynamic)
	model.RoutingPolicy = types.StringValue(record.RoutingPolicy)
	model.SyncStatus = types.StringValue(record.SyncStatus)
	model.FQDN = types.StringValue(fqdn)

	// Map values
	model.Values = make([]RecordValueDataModel, 0, len(record.Values))
	for _, v := range record.Values {
		valueModel := RecordValueDataModel{
			Value:    types.StringValue(v.Value),
			Priority: types.Int64Value(int64(v.Priority)),
			Weight:   types.Int64Value(int64(v.Weight)),
		}
		if v.MonitorID != nil {
			valueModel.MonitorID = types.StringValue(*v.MonitorID)
		} else {
			valueModel.MonitorID = types.StringNull()
		}
		model.Values = append(model.Values, valueModel)
	}
}

// =============================================================================
// Records Data Source (plural - list all records in a zone)
// =============================================================================

// Ensure RecordsDataSource satisfies the datasource.DataSource interface
var (
	_ datasource.DataSource              = &RecordsDataSource{}
	_ datasource.DataSourceWithConfigure = &RecordsDataSource{}
)

// RecordsDataSource defines the records data source implementation
type RecordsDataSource struct {
	client *client.Client
}

// RecordsDataSourceModel describes the records data source data model
type RecordsDataSourceModel struct {
	ZoneID  types.String            `tfsdk:"zone_id"`
	Records []RecordDataSourceModel `tfsdk:"records"`
}

// NewRecordsDataSource creates a new records data source
func NewRecordsDataSource() datasource.DataSource {
	return &RecordsDataSource{}
}

// Metadata returns the data source type name
func (d *RecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

// Schema defines the schema for the records data source
func (d *RecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all DNS records in a zone.",
		MarkdownDescription: "Fetches all DNS records in a zone in NorainaDNS.",
		Attributes: map[string]schema.Attribute{
			"zone_id": schema.StringAttribute{
				Description: "The ID of the zone to fetch records from.",
				Required:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "List of records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The record ID.",
							Computed:    true,
						},
						"zone_id": schema.StringAttribute{
							Description: "The zone ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The record name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The record type.",
							Computed:    true,
						},
						"ttl": schema.Int64Attribute{
							Description: "The TTL.",
							Computed:    true,
						},
						"is_dynamic": schema.BoolAttribute{
							Description: "Is dynamic.",
							Computed:    true,
						},
						"routing_policy": schema.StringAttribute{
							Description: "Routing policy.",
							Computed:    true,
						},
						"sync_status": schema.StringAttribute{
							Description: "Sync status.",
							Computed:    true,
						},
						"fqdn": schema.StringAttribute{
							Description: "The FQDN.",
							Computed:    true,
						},
						"values": schema.ListNestedAttribute{
							Description: "The values.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"value": schema.StringAttribute{
										Description: "The value.",
										Computed:    true,
									},
									"priority": schema.Int64Attribute{
										Description: "Priority.",
										Computed:    true,
									},
									"weight": schema.Int64Attribute{
										Description: "Weight.",
										Computed:    true,
									},
									"monitor_id": schema.StringAttribute{
										Description: "Monitor ID.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client
func (d *RecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read reads the records data source
func (d *RecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RecordsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID := config.ZoneID.ValueString()

	tflog.Debug(ctx, "Reading records data source", map[string]interface{}{
		"zone_id": zoneID,
	})

	records, err := d.client.GetRecords(ctx, zoneID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Records",
			fmt.Sprintf("Could not read records for zone %s: %s", zoneID, err.Error()),
		)
		return
	}

	// Get zone info for FQDN
	zone, _ := d.client.GetZone(ctx, zoneID)

	var state RecordsDataSourceModel
	state.ZoneID = config.ZoneID
	for _, record := range records {
		var recordModel RecordDataSourceModel
		fqdn := d.buildFQDN(record.Name, zone)
		d.mapRecordToModelPtr(&record, &recordModel, fqdn)
		state.Records = append(state.Records, recordModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// buildFQDN builds the fully qualified domain name
func (d *RecordsDataSource) buildFQDN(name string, zone *client.Zone) string {
	if zone == nil {
		return ""
	}
	if name == "@" {
		return zone.DomainName + "."
	}
	return name + "." + zone.DomainName + "."
}

// mapRecordToModelPtr maps a Record API response to the data source model
func (d *RecordsDataSource) mapRecordToModelPtr(record *client.Record, model *RecordDataSourceModel, fqdn string) {
	model.ID = types.StringValue(record.ID)
	model.ZoneID = types.StringValue(record.ZoneID)
	model.Name = types.StringValue(record.Name)
	model.Type = types.StringValue(record.RecordType)
	model.TTL = types.Int64Value(record.TTL)
	model.IsDynamic = types.BoolValue(record.IsDynamic)
	model.RoutingPolicy = types.StringValue(record.RoutingPolicy)
	model.SyncStatus = types.StringValue(record.SyncStatus)
	model.FQDN = types.StringValue(fqdn)

	// Map values
	model.Values = make([]RecordValueDataModel, 0, len(record.Values))
	for _, v := range record.Values {
		valueModel := RecordValueDataModel{
			Value:    types.StringValue(v.Value),
			Priority: types.Int64Value(int64(v.Priority)),
			Weight:   types.Int64Value(int64(v.Weight)),
		}
		if v.MonitorID != nil {
			valueModel.MonitorID = types.StringValue(*v.MonitorID)
		} else {
			valueModel.MonitorID = types.StringNull()
		}
		model.Values = append(model.Values, valueModel)
	}
}
