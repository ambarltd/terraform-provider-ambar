// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	Ambar "github.com/ambarltd/ambar_go_client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &dataSourceResource{}
var _ resource.ResourceWithImportState = &dataSourceResource{}
var _ resource.ResourceWithConfigure = &dataSourceResource{}

func NewDataSourceResource() resource.Resource {
	return &dataSourceResource{}
}

// dataSourceResource defines the resource implementation.
type dataSourceResource struct {
	client *Ambar.APIClient
}

// dataSourceResourceModel describes the resource data model.
type dataSourceResourceModel struct {
	DataSourceType   types.String `tfsdk:"data_source_type"`
	Description      types.String `tfsdk:"description"`
	DataSourceConfig types.Map    `tfsdk:"data_source_config"`
	State            types.String `tfsdk:"state"`
	ResourceId       types.String `tfsdk:"resource_id"`
}

func (r *dataSourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (r *dataSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ambar DataSource resource. Represents the details needed for Ambar to establish a connection to your database storage which is then used to import record sequences into Ambar.",
		Description:         "Ambar DataSource resource. Represents the details needed for Ambar to establish a connection to your database storage which is then used to import record sequences into Ambar.",

		Attributes: map[string]schema.Attribute{
			"data_source_type": schema.StringAttribute{
				MarkdownDescription: "The type of durable storage being connected to. This should be one of the supported database types by Ambar such as postgres. See Ambar documentation for a full list of supported data_source_types.",
				Description:         "The type of durable storage being connected to. This should be one of the supported database types by Ambar such as postgres. See Ambar documentation for a full list of supported data_source_types.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A user friendly description of this DataSource. Use the description field to help augment information about this DataSource which may not be apparent from describing the resource, such as if it is a test environment resource or which department owns it.",
				Description:         "A user friendly description of this DataSource. Use the description field to help augment information about this DataSource which may not be apparent from describing the resource, such as if it is a test environment resource or which department owns it.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_source_config": schema.MapAttribute{
				MarkdownDescription: "A Key Value map of further DataSource configurations specific to the type of database this DataSource will connect to. See Ambar documentation for a list of required parameters.",
				Description:         "A Key Value map of further DataSource configurations specific to the type of database this DataSource will connect to. See Ambar documentation for a list of required parameters.",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
							var current dataSourceResourceModel
							// Read Terraform prior state data into the model
							resp.Diagnostics.Append(req.State.Get(ctx, &current)...)

							if resp.Diagnostics.HasError() {
								return
							}

							var plan dataSourceResourceModel
							// Read Terraform plan data into the model
							resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

							if resp.Diagnostics.HasError() {
								return
							}

							// Loop through all the values and compare. If anything other than username / password has changed we will need to
							// flag to replace the resource.
							for key, value := range plan.DataSourceConfig.Elements() {
								switch key {
								case
									"username",
									"password",
									"hostname",
									"hostPort",
									"tlsTerminationOverrideHost":
									tflog.Info(ctx, "Detected change in config which does not require replace")
									continue
								default:
									if value.String() != current.DataSourceConfig.Elements()[key].String() {
										tflog.Info(ctx, "Detected change in config which requires replace")
										resp.RequiresReplace = true
										return
									}
								}
							}
						},
						"Only some config value updates are supported. For a list of supported values, see Ambar docs.",
						"Only some config value updates are supported. For a list of supported values, see Ambar docs.",
					),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the Ambar resource.",
				Description:         "The current state of the Ambar resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(doesStateRequireReplace,
						"If the state of the resource is not valid for further use, such as FAILED",
						"If the state of the resource is not valid for further use, such as FAILED"),
				},
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "The unique Ambar resource id for this resource.",
				Description:         "The unique Ambar resource id for this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func doesStateRequireReplace(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var data dataSourceResourceModel

	// Read Terraform prior state data into the model
	request.State.Get(ctx, &data)

	tflog.Debug(ctx, "Got value for state: "+data.State.String())

	if data.State.String() == "FAILED" {
		response.RequiresReplace = true
		return
	}
}

func (r *dataSourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Ambar.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Ambar.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *dataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan dataSourceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createDataSource Ambar.CreateDataSourceRequest
	createDataSource.DataSourceType = plan.DataSourceType.ValueString()
	createDataSource.Description = plan.Description.ValueStringPointer()

	// Handle dynamic DataSource resource configuration map
	createDataSource.DataSourceConfig = make(map[string]string)
	for key, value := range plan.DataSourceConfig.Elements() {
		// Remove the quotes if any are present.
		createDataSource.DataSourceConfig[key] = strings.Trim(value.String(), "\"")
	}

	// Create the API call and execute it
	createResourceResponse, httpResponse, err := r.client.AmbarAPI.CreateDataSource(ctx).CreateDataSourceRequest(createDataSource).Execute()
	if err != nil || createResourceResponse == nil || httpResponse == nil {
		tflog.Debug(ctx, "StatusCode: "+httpResponse.Status)
		httpBody, _ := io.ReadAll(httpResponse.Body)
		errString := string(httpBody)
		tflog.Debug(ctx, errString)
		resp.Diagnostics.AddError(
			"Error creating DataSource",
			"Could not create DataSource: "+AmbarApiErrorToTerraformErrorString(errString),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ResourceId = types.StringValue(createResourceResponse.ResourceId)
	plan.State = types.StringValue(createResourceResponse.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = createResourceResponse.ResourceId

	var describeResourceResponse *Ambar.DataSource

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, _, err = r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
		if err != nil {
			tflog.Error(ctx, "Got error!"+err.Error())
			return
		}

		tflog.Debug(ctx, "Got state: "+describeResourceResponse.State)
		if describeResourceResponse.State == "READY" {
			break
		}

		if describeResourceResponse.State == "FAILED" {
			tflog.Info(ctx, "Error creating the DataSource, failing creation.")
			resp.Diagnostics.AddError(
				"Error creating DataSource",
				"Could not create DataSource, resource in FAILED state indicating errors while creating with passed values.",
			)
			return
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan.State = types.StringValue(describeResourceResponse.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataSourceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the latest state from the Ambar describe API
	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = data.ResourceId.ValueString()

	describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()

	if err != nil {
		tflog.Error(ctx, "Got error: "+err.Error())

		if httpResponse.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "Resource was not found. Removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Error(ctx, "Unexpected error, dumping logs and returning.")
		tflog.Error(ctx, "Got http response, code: "+strconv.Itoa(httpResponse.StatusCode)+", status: "+httpResponse.Status)
		resp.Diagnostics.AddError("Unable to read DataSource resource.", err.Error())
		return
	}

	tflog.Debug(ctx, "Got state: "+describeResourceResponse.State)
	// If the resource is in the deleting state, then we should consider it deleted.
	if describeResourceResponse.State == "DELETING" {
		tflog.Info(ctx, "Resource was found in DELETING state and will not exist eventually. Removing from state.")
		resp.State.RemoveResource(ctx)
		return
	}

	// Ambar resources are immutable except for state changes when resources are creating / updating / deleting
	// Where updates would be service side changes (updating customer infra, system maintenance, etc)
	// we will still do data updates here inorder to support imports, though they will only be partial as API's do
	// not return some sensitive data like credential information
	data.State = types.StringValue(describeResourceResponse.State)
	data.DataSourceType = types.StringValue(describeResourceResponse.DataSourceType)
	data.Description = types.StringPointerValue(describeResourceResponse.Description)

	// Describe calls will not return sensitive credentials. So we will need to carry the local value forward to prevent
	// always doing a replacement on each apply.
	username := data.DataSourceConfig.Elements()["username"].String()
	password := data.DataSourceConfig.Elements()["password"].String()

	tflog.Info(ctx, "Got value for username from state: "+username)

	// Add back credentials to prevent recreation issues.
	describeResourceResponse.DataSourceConfig["username"] = strings.Trim(username, "\"")
	describeResourceResponse.DataSourceConfig["password"] = strings.Trim(password, "\"")

	// remap the config from the describe call. This will be missing credentials
	data.DataSourceConfig, _ = types.MapValueFrom(ctx, types.StringType, describeResourceResponse.DataSourceConfig)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Ambar does not support resource updates on DataSources for now.
	var plan dataSourceResourceModel
	var current dataSourceResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &current)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// We need to validate we can perform the change in a single operation, so only credentials should be changed, or only
	// non-credential attributes should be changed
	var credentialsUpdated = plan.DataSourceConfig.Elements()["username"].String() != current.DataSourceConfig.Elements()["username"].String() ||
		plan.DataSourceConfig.Elements()["password"].String() != current.DataSourceConfig.Elements()["password"].String()
	var nonCredentialsUpdated = plan.DataSourceConfig.Elements()["hostname"].String() != current.DataSourceConfig.Elements()["hostname"].String() ||
		plan.DataSourceConfig.Elements()["hostPort"].String() != current.DataSourceConfig.Elements()["hostPort"].String()

	if _, ok := plan.DataSourceConfig.Elements()["tlsTerminationOverrideHost"]; ok {
		nonCredentialsUpdated = nonCredentialsUpdated ||
			plan.DataSourceConfig.Elements()["tlsTerminationOverrideHost"].String() != current.DataSourceConfig.Elements()["tlsTerminationOverrideHost"].String()
	}

	var updateResourceResponse Ambar.ResourceStateChangeResponse

	if credentialsUpdated {
		// Make the call to update the credentials if requested
		var updateCredentialsRequest Ambar.UpdateResourceCredentialsRequest
		updateCredentialsRequest.ResourceId = plan.ResourceId.ValueString()
		updateCredentialsRequest.Username = strings.Trim(plan.DataSourceConfig.Elements()["username"].String(), "\"")
		updateCredentialsRequest.Password = strings.Trim(plan.DataSourceConfig.Elements()["password"].String(), "\"")

		updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataSourceCredentials(ctx).UpdateResourceCredentialsRequest(updateCredentialsRequest).Execute()
		if err != nil || updateResourceResponse == nil || httpResponse == nil {
			resp.Diagnostics.AddError(
				"Error updating DataSource",
				"Could not update DataSource, unexpected error: "+err.Error(),
			)
			return
		}

		r.waitSourceForResourceReady(plan.ResourceId.ValueString(), ctx)
	}

	if nonCredentialsUpdated {
		// Make the call to update the DataSource attributes
		var updateDataSourceRequest Ambar.UpdateDataSourceRequest
		updateDataSourceRequest.ResourceId = plan.ResourceId.ValueString()
		var port = strings.Trim(plan.DataSourceConfig.Elements()["hostPort"].String(), "\"")
		updateDataSourceRequest.Port = &port
		var hostname = strings.Trim(plan.DataSourceConfig.Elements()["hostname"].String(), "\"")
		updateDataSourceRequest.Hostname = &hostname

		if _, ok := plan.DataSourceConfig.Elements()["tlsTerminationOverrideHost"]; ok {
			var tlsHost = strings.Trim(plan.DataSourceConfig.Elements()["tlsTerminationOverrideHost"].String(), "\"")
			updateDataSourceRequest.TlsTerminationOverrideHost = &tlsHost
		}

		updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataSource(ctx).UpdateDataSourceRequest(updateDataSourceRequest).Execute()
		if err != nil || updateResourceResponse == nil || httpResponse == nil {
			resp.Diagnostics.AddError(
				"Error updating DataSource",
				"Could not update DataSource, unexpected error: "+err.Error(),
			)
			return
		}

		r.waitSourceForResourceReady(plan.ResourceId.ValueString(), ctx)
	}

	// partial state save in case of interrupt
	plan.State = types.StringValue(updateResourceResponse.State)

	// state save in case of interrupt
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataSourceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the DataSource
	var deleteDataSource Ambar.DeleteResourceRequest
	deleteDataSource.ResourceId = data.ResourceId.ValueString()

	deleteResponse, httpResponse, err := r.client.AmbarAPI.DeleteDataSource(ctx).DeleteResourceRequest(deleteDataSource).Execute()
	if err != nil {
		tflog.Error(ctx, "Got error: "+err.Error())

		if httpResponse.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "Resource was not found. Removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Error(ctx, "Unexpected error, dumping logs and returning.")
		tflog.Error(ctx, "Got http response, code: "+strconv.Itoa(httpResponse.StatusCode)+", status: "+httpResponse.Status)
		resp.Diagnostics.AddError("Unable to delete DataSource resource.", err.Error())
		return
	}
	tflog.Info(ctx, "Got deleteResponse: "+deleteResponse.State)

	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()

		if err != nil {
			tflog.Error(ctx, "Got error: "+err.Error())

			if httpResponse.StatusCode == http.StatusNotFound {
				tflog.Info(ctx, "Resource was not found. This is expected for delete, returning.")
				return
			}

			tflog.Error(ctx, "Unexpected error, dumping logs and returning.")
			tflog.Error(ctx, "Got http response, code: "+strconv.Itoa(httpResponse.StatusCode)+", status: "+httpResponse.Status)
			resp.Diagnostics.AddError("Unable to read DataSource resource to confirm deletion. Got error.", err.Error())
			return
		}

		tflog.Debug(ctx, "Waiting for resource to complete deletion. Current state: "+describeResourceResponse.State)
	}
}

func (r *dataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *dataSourceResource) waitSourceForResourceReady(resourceId string, ctx context.Context) {
	// Wait for the update to complete

	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = resourceId

	for {
		time.Sleep(10 * time.Second)
		describeResourceResponse, _, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
		if err != nil {
			return
		}

		if describeResourceResponse.State == "READY" {
			break
		}
	}
}
