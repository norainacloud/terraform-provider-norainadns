// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/norainacloud/terraform-provider-norainadns/internal/client"
)

// Ensure RecordResource satisfies the resource.Resource interface
var (
	_ resource.Resource                = &RecordResource{}
	_ resource.ResourceWithConfigure   = &RecordResource{}
	_ resource.ResourceWithImportState = &RecordResource{}
)

// RecordResource defines the record resource implementation
type RecordResource struct {
	client *client.Client
}

// RecordResourceModel describes the record resource data model
type RecordResourceModel struct {
	ID            types.String       `tfsdk:"id"`
	ZoneID        types.String       `tfsdk:"zone_id"`
	Name          types.String       `tfsdk:"name"`
	Type          types.String       `tfsdk:"type"`
	TTL           types.Int64        `tfsdk:"ttl"`
	IsDynamic     types.Bool         `tfsdk:"is_dynamic"`
	RoutingPolicy types.String       `tfsdk:"routing_policy"`
	Values        []RecordValueModel `tfsdk:"values"`
	SyncStatus    types.String       `tfsdk:"sync_status"`
	FQDN          types.String       `tfsdk:"fqdn"`
}

// RecordValueModel describes a record value
type RecordValueModel struct {
	Value     types.String `tfsdk:"value"`
	Priority  types.Int64  `tfsdk:"priority"`
	Weight    types.Int64  `tfsdk:"weight"`
	MonitorID types.String `tfsdk:"monitor_id"`
}

// NewRecordResource creates a new record resource
func NewRecordResource() resource.Resource {
	return &RecordResource{}
}

// Metadata returns the resource type name
func (r *RecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

// Schema defines the schema for the record resource
func (r *RecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS record in NorainaDNS.",
		MarkdownDescription: `Manages a DNS record in NorainaDNS.

Records are automatically synchronized to both AWS Route53 and GCP Cloud DNS.

## Failover Routing

For automatic failover, set ` + "`routing_policy`" + ` to ` + "`failover`" + ` and provide two values:
- A primary value (priority 0) with a ` + "`monitor_id`" + ` for health checking
- A secondary/backup value (priority 1) that will be used when the primary is unhealthy

~> **Important:** When using failover routing, the monitor must exist before creating the record.
Use ` + "`depends_on`" + ` or create the monitor resource first.

## Example with Failover

` + "```hcl" + `
resource "norainadns_monitor" "origin" {
  name        = "Origin Server"
  protocol    = "HTTPS"
  target_host = "origin.example.com"
  target_path = "/health"
}

resource "norainadns_record" "www" {
  zone_id        = norainadns_zone.example.id
  name           = "www"
  type           = "A"
  ttl            = 60
  routing_policy = "failover"

  values {
    value      = "192.0.2.10"  # Primary (CDN)
    priority   = 0
    monitor_id = norainadns_monitor.origin.id
  }

  values {
    value    = "192.0.2.20"  # Backup (Origin)
    priority = 1
  }
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this record.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_id": schema.StringAttribute{
				Description: "The ID of the zone this record belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "The name of the record (subdomain). Use '@' for the root domain.",
				MarkdownDescription: "The name of the record (subdomain). Use `@` for the root domain.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				Description: "The record type (A, AAAA, CNAME, TXT, MX, NS, SRV, CAA, PTR).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA", "PTR"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds. Defaults to 300.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300),
				Validators: []validator.Int64{
					int64validator.Between(60, 86400),
				},
			},
			"is_dynamic": schema.BoolAttribute{
				Description:         "Whether this is a dynamic record with health checking.",
				MarkdownDescription: "Whether this is a dynamic record with health checking. Set to `true` for failover routing.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"routing_policy": schema.StringAttribute{
				Description: "The routing policy (simple, failover, weighted, latency). Defaults to 'simple'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("simple"),
				Validators: []validator.String{
					stringvalidator.OneOf("simple", "failover", "weighted", "latency"),
				},
			},
			"sync_status": schema.StringAttribute{
				Description: "The synchronization status of this record.",
				Computed:    true,
			},
			"fqdn": schema.StringAttribute{
				Description: "The fully qualified domain name for this record.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"values": schema.ListNestedBlock{
				Description:         "The values for this record. At least one value is required.",
				MarkdownDescription: "The values for this record. At least one value is required. For failover, provide two values with priorities 0 (primary) and 1 (secondary).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Description: "The record value (e.g., IP address for A records).",
							Required:    true,
						},
						"priority": schema.Int64Attribute{
							Description:         "The priority of this value. 0 = primary, 1 = secondary.",
							MarkdownDescription: "The priority of this value. `0` = primary, `1` = secondary. Lower is higher priority.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
						},
						"weight": schema.Int64Attribute{
							Description: "The weight for weighted routing. Assigned by the server when not set.",
							Optional:    true,
							Computed:    true,
						},
						"monitor_id": schema.StringAttribute{
							Description:         "The ID of the monitor for health-checked failover.",
							MarkdownDescription: "The ID of the monitor for health-checked failover. Required for the primary value when using `failover` routing.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client
func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new record
func (r *RecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID := plan.ZoneID.ValueString()

	tflog.Debug(ctx, "Creating record", map[string]interface{}{
		"zone_id": zoneID,
		"name":    plan.Name.ValueString(),
		"type":    plan.Type.ValueString(),
	})

	// Build the create request
	createReq := client.CreateRecordRequest{
		Name:          plan.Name.ValueString(),
		RecordType:    plan.Type.ValueString(),
		TTL:           plan.TTL.ValueInt64(),
		IsDynamic:     plan.IsDynamic.ValueBool(),
		RoutingPolicy: plan.RoutingPolicy.ValueString(),
		Values:        make([]client.CreateRecordValueRequest, 0, len(plan.Values)),
	}

	for _, v := range plan.Values {
		valueReq := client.CreateRecordValueRequest{
			Value:    v.Value.ValueString(),
			Priority: int(v.Priority.ValueInt64()),
			Weight:   int(v.Weight.ValueInt64()),
		}
		if !v.MonitorID.IsNull() && !v.MonitorID.IsUnknown() {
			monitorID := v.MonitorID.ValueString()
			valueReq.MonitorID = &monitorID
		}
		createReq.Values = append(createReq.Values, valueReq)
	}

	record, err := r.client.CreateRecord(ctx, zoneID, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Record",
			fmt.Sprintf("Could not create record %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Get zone info for FQDN
	zone, _ := r.client.GetZone(ctx, zoneID)
	fqdn := r.buildFQDN(plan.Name.ValueString(), zone)

	// Map response to state
	r.mapRecordToModel(record, &plan, fqdn)

	tflog.Debug(ctx, "Created record", map[string]interface{}{
		"id": record.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the current state of a record
func (r *RecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID, recordID, err := r.parseCompositeID(state.ID.ValueString(), state.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing IDs", err.Error())
		return
	}

	tflog.Debug(ctx, "Reading record", map[string]interface{}{
		"zone_id":   zoneID,
		"record_id": recordID,
	})

	record, err := r.client.GetRecord(ctx, zoneID, recordID)
	if err != nil {
		// If the record doesn't exist, remove it from state
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "RECORD_NOT_FOUND" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Record",
			fmt.Sprintf("Could not read record ID %s: %s", recordID, err.Error()),
		)
		return
	}

	// Get zone info for FQDN
	zone, _ := r.client.GetZone(ctx, zoneID)
	fqdn := r.buildFQDN(record.Name, zone)

	// Map response to state
	r.mapRecordToModel(record, &state, fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates an existing record
func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID, recordID, err := r.parseCompositeID(plan.ID.ValueString(), plan.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing IDs", err.Error())
		return
	}

	tflog.Debug(ctx, "Updating record", map[string]interface{}{
		"zone_id":   zoneID,
		"record_id": recordID,
	})

	// Build the update request
	isDynamic := plan.IsDynamic.ValueBool()
	updateReq := client.UpdateRecordRequest{
		Name:          plan.Name.ValueString(),
		TTL:           plan.TTL.ValueInt64(),
		IsDynamic:     &isDynamic,
		RoutingPolicy: plan.RoutingPolicy.ValueString(),
		Values:        make([]client.CreateRecordValueRequest, 0, len(plan.Values)),
	}

	for _, v := range plan.Values {
		valueReq := client.CreateRecordValueRequest{
			Value:    v.Value.ValueString(),
			Priority: int(v.Priority.ValueInt64()),
			Weight:   int(v.Weight.ValueInt64()),
		}
		if !v.MonitorID.IsNull() && !v.MonitorID.IsUnknown() {
			monitorID := v.MonitorID.ValueString()
			valueReq.MonitorID = &monitorID
		}
		updateReq.Values = append(updateReq.Values, valueReq)
	}

	record, err := r.client.UpdateRecord(ctx, zoneID, recordID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Record",
			fmt.Sprintf("Could not update record ID %s: %s", recordID, err.Error()),
		)
		return
	}

	// Get zone info for FQDN
	zone, _ := r.client.GetZone(ctx, zoneID)
	fqdn := r.buildFQDN(record.Name, zone)

	// Map response to state
	r.mapRecordToModel(record, &plan, fqdn)

	tflog.Debug(ctx, "Updated record", map[string]interface{}{
		"id": record.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes a record
func (r *RecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneID, recordID, err := r.parseCompositeID(state.ID.ValueString(), state.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing IDs", err.Error())
		return
	}

	tflog.Debug(ctx, "Deleting record", map[string]interface{}{
		"zone_id":   zoneID,
		"record_id": recordID,
	})

	err = r.client.DeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		// If the record is already deleted, that's fine
		if apiErr, ok := err.(*client.APIError); ok && apiErr.Code == "RECORD_NOT_FOUND" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Record",
			fmt.Sprintf("Could not delete record ID %s: %s", recordID, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleted record", map[string]interface{}{
		"record_id": recordID,
	})
}

// ImportState imports an existing record into Terraform state
// Format: zone_id/record_id (e.g., "550e8400-e29b-41d4-a716-446655440000/7c9e6679-7425-40de-944b-e07fc1f90ae7")
func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: zone_id/record_id (e.g., '550e8400-e29b-41d4-a716-446655440000/7c9e6679-7425-40de-944b-e07fc1f90ae7'), got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// parseCompositeID parses zone_id and record_id from state
func (r *RecordResource) parseCompositeID(recordIDStr, zoneIDStr string) (zoneID, recordID string, err error) {
	return zoneIDStr, recordIDStr, nil
}

// buildFQDN builds the fully qualified domain name
func (r *RecordResource) buildFQDN(name string, zone *client.Zone) string {
	if zone == nil {
		return ""
	}
	if name == "@" {
		return zone.DomainName + "."
	}
	return name + "." + zone.DomainName + "."
}

// mapRecordToModel maps a Record API response to the Terraform model
func (r *RecordResource) mapRecordToModel(record *client.Record, model *RecordResourceModel, fqdn string) {
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
	model.Values = make([]RecordValueModel, 0, len(record.Values))
	for _, v := range record.Values {
		valueModel := RecordValueModel{
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
