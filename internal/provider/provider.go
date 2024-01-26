// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"

	Ambar "github.com/ambarltd/ambar_go_client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ambarProvider satisfies various provider interfaces.
var _ provider.Provider = &ambarProvider{}

// ambarProvider defines the provider implementation.
type ambarProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ambarProviderModel describes the provider data model.
type ambarProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Api_key  types.String `tfsdk:"api_key"`
}

func (p *ambarProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ambar"
	resp.Version = p.version
}

func (p *ambarProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with your regional Ambar environment.",
		Description:         "Interact with your regional Ambar environment.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The Ambar API URI to use for these resources. Note that Ambar has region specific endpoints, so be sure to set this to the region your key was created in. May also be provided via the AMBAR_ENDPOINT environment variable",
				Description:         "The Ambar API URI to use for these resources. Note that Ambar has region specific endpoints, so be sure to set this to the region your key was created in. May also be provided via the AMBAR_ENDPOINT environment variable",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API Key for your Ambar environment. Keys are region specific, so make sure to use a key which is valid for the selected Ambar endpoint. May also be provided via the AMBAR_ENVIRONMENT_KEY environment variable",
				Description:         "The API Key for your Ambar environment. Keys are region specific, so make sure to use a key which is valid for the selected Ambar endpoint. May also be provided via the AMBAR_ENVIRONMENT_KEY environment variable",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ambarProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Ambar client")
	// Retrieve provider data from configuration
	var config ambarProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Ambar endpoint.",
			"The provider cannot create the Ambar API client as there is an unknown configuration value for the Ambar API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AMBAR_ENDPOINT environment variable.",
		)
	}

	if config.Api_key.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Ambar API Key",
			"The provider cannot create the Ambar API client as there is an unknown configuration value for the Ambar API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AMBAR_ENVIRONMENT_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	endpoint := os.Getenv("AMBAR_ENDPOINT")
	api_key := os.Getenv("AMBAR_ENVIRONMENT_KEY")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Api_key.IsNull() {
		api_key = config.Api_key.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Ambar API endpoint",
			"The provider cannot create the Ambar API client as there is a missing or empty value for the Ambar API endpoint. "+
				"Set the endpoint value in the configuration or use the AMBAR_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Ambar API Key",
			"The provider cannot create the Ambar API client as there is a missing or empty value for the Ambar API Key. "+
				"Set the key value in the configuration or use the AMBAR_ENVIRONMENT_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "ambar_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "ambar_environment_key", api_key)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "ambar_environment_key")
	tflog.Info(ctx, "Creating Ambar client")

	cfg := Ambar.NewConfiguration()
	cfg.AddDefaultHeader("x-api-key", api_key)
	cfg.Host = endpoint

	client := Ambar.NewAPIClient(cfg)

	// Make the Ambar client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured Ambar client", map[string]any{"success": true})
}

func (p *ambarProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDataSourceResource,
		NewFilterResource,
		NewDataDestinationResource,
	}
}

// DataSources Ambar does not currently have any *Terraform* Data Source resource types, not to be confused with the Ambar DataSource resource type.
func (p *ambarProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ambarProvider{
			version: version,
		}
	}
}
