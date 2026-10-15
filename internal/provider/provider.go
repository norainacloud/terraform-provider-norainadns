// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the NorainaDNS Terraform provider
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/norainacloud/terraform-provider-norainadns/internal/client"
)

// Ensure NorainaDNSProvider satisfies the provider.Provider interface
var _ provider.Provider = &NorainaDNSProvider{}

// NorainaDNSProvider defines the provider implementation
type NorainaDNSProvider struct {
	version string
}

// NorainaDNSProviderModel describes the provider data model
type NorainaDNSProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

// New returns a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NorainaDNSProvider{
			version: version,
		}
	}
}

// Metadata returns the provider metadata
func (p *NorainaDNSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "norainadns"
	resp.Version = p.version
}

// Schema defines the provider configuration schema
func (p *NorainaDNSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The NorainaDNS provider allows you to manage DNS zones, records, and health monitors in NorainaDNS.",
		MarkdownDescription: `
The NorainaDNS provider allows you to manage DNS zones, records, and health monitors in NorainaDNS.

NorainaDNS is a multi-tenant DNS management system that synchronizes zones and records across both
multiple DNS servers, providing automatic failover support via health monitoring.

## Authentication

The provider requires an API key for authentication. You can generate an API key from the NorainaDNS
web interface under Settings > API Keys.

The API key can be provided in two ways:

1. Via the provider configuration:
` + "```hcl" + `
provider "norainadns" {
  endpoint = "https://dash.norainadns.com"
  api_key  = "nordns_your_api_key_here"
}
` + "```" + `

2. Via environment variables:
` + "```bash" + `
export NORAINADNS_ENDPOINT="https://dash.norainadns.com"
export NORAINADNS_API_KEY="nordns_your_api_key_here"
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description:         "The NorainaDNS API endpoint URL. Can also be set via NORAINADNS_ENDPOINT environment variable.",
				MarkdownDescription: "The NorainaDNS API endpoint URL. Can also be set via `NORAINADNS_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				Description:         "The API key for authenticating with NorainaDNS. Can also be set via NORAINADNS_API_KEY environment variable.",
				MarkdownDescription: "The API key for authenticating with NorainaDNS. Can also be set via `NORAINADNS_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure configures the provider with the given configuration
func (p *NorainaDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring NorainaDNS provider")

	var config NorainaDNSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for unknown values
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown NorainaDNS API Endpoint",
			"The provider cannot create the NorainaDNS API client as there is an unknown configuration value for the NorainaDNS API endpoint. "+
				"Either set the value statically in the configuration, or use the NORAINADNS_ENDPOINT environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown NorainaDNS API Key",
			"The provider cannot create the NorainaDNS API client as there is an unknown configuration value for the NorainaDNS API key. "+
				"Either set the value statically in the configuration, or use the NORAINADNS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from config or environment
	endpoint := os.Getenv("NORAINADNS_ENDPOINT")
	apiKey := os.Getenv("NORAINADNS_API_KEY")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// Validate required values
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing NorainaDNS API Endpoint",
			"The provider cannot create the NorainaDNS API client as there is a missing configuration value for the NorainaDNS API endpoint. "+
				"Set the endpoint value in the configuration or use the NORAINADNS_ENDPOINT environment variable.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing NorainaDNS API Key",
			"The provider cannot create the NorainaDNS API client as there is a missing configuration value for the NorainaDNS API key. "+
				"Set the api_key value in the configuration or use the NORAINADNS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the API client
	tflog.Debug(ctx, "Creating NorainaDNS API client", map[string]interface{}{
		"endpoint": endpoint,
	})

	apiClient := client.NewClient(client.ClientConfig{
		BaseURL: endpoint,
		APIKey:  apiKey,
	})

	// Make the client available to resources and data sources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	tflog.Info(ctx, "Configured NorainaDNS provider", map[string]interface{}{
		"endpoint": endpoint,
	})
}

// Resources returns the list of resources supported by this provider
func (p *NorainaDNSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewZoneResource,
		NewMonitorResource,
		NewRecordResource,
	}
}

// DataSources returns the list of data sources supported by this provider
func (p *NorainaDNSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewZoneDataSource,
		NewZonesDataSource,
		NewMonitorDataSource,
		NewMonitorsDataSource,
		NewRecordDataSource,
		NewRecordsDataSource,
	}
}
