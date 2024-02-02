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
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &dataSourceResource{}
var _ resource.ResourceWithImportState = &dataSourceResource{}

func NewDataSourceResource() resource.Resource {
	return &dataSourceResource{}
}

// dataSourceResource defines the resource implementation.
type dataSourceResource struct {
	client *Ambar.APIClient
}

// dataSourceResourceModel describes the resource data model.
type dataSourceResourceModel struct {
	DataSourceType     types.String `tfsdk:"data_source_type"`
	Description        types.String `tfsdk:"description"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	SerialColumn       types.String `tfsdk:"serial_column"`
	PartitioningColumn types.String `tfsdk:"partitioning_column"`
	DataSourceConfig   types.Map    `tfsdk:"data_source_config"`
	State              types.String `tfsdk:"state"`
	ResourceId         types.String `tfsdk:"resource_id"`
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
			"username": schema.StringAttribute{
				MarkdownDescription: "A username credential which Ambar can use to communicate with your database storage.",
				Description:         "A username credential which Ambar can use to communicate with your database storage.",
				Required:            true,
				Sensitive:           true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "A password credential which Ambar can use to communicate with your database storage.",
				Description:         "A password credential which Ambar can use to communicate with your database storage.",
				Required:            true,
				Sensitive:           true,
			},
			"serial_column": schema.StringAttribute{
				MarkdownDescription: "The name of a column which increments with each write to the database.",
				Description:         "The name of a column which increments with each write to the database.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"partitioning_column": schema.StringAttribute{
				MarkdownDescription: "The name of the column which records in the database are partitioned on.",
				Description:         "The name of the column which records in the database are partitioned on.",
				Required:            true,
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
					mapplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the Ambar resource.",
				Description:         "The current state of the Ambar resource.",
				Computed:            true,
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
	createDataSource.PartitioningColumn = plan.PartitioningColumn.ValueString()
	createDataSource.SerialColumn = plan.SerialColumn.ValueString()
	createDataSource.Username = plan.Username.ValueString()
	createDataSource.Password = plan.Password.ValueString()

	// Handle dynamic DataSource resource configuration map
	createDataSource.DataSourceConfig = make(map[string]string)
	for key, value := range plan.DataSourceConfig.Elements() {
		// Remove the quotes if any are present.
		createDataSource.DataSourceConfig[key] = strings.Trim(value.String(), "\"")
	}

	// Create the API call and execute it
	createResourceResponse, httpResponse, err := r.client.AmbarAPI.CreateDataSource(ctx).CreateDataSourceRequest(createDataSource).Execute()
	if err != nil || createResourceResponse == nil || httpResponse == nil {
		resp.Diagnostics.AddError(
			"Error creating DataSource",
			"Could not create DataSource, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ResourceId = types.StringValue(createResourceResponse.ResourceId)
	plan.State = types.StringValue(createResourceResponse.ResourceState)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = createResourceResponse.ResourceId

	var describeResourceResponse *Ambar.DataSource

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, _, err = r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
		if err != nil {
			return
		}

		if describeResourceResponse.State == "READY" {
			break
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan.State = types.StringValue(describeResourceResponse.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
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

	describeResourceResponse, _, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
	// Todo: Handle ResourceNotFoundException gracefully per https://developer.hashicorp.com/terraform/plugin/framework/resources/read#recommendations
	if err != nil {
		return
	}

	// Ambar resources are immutable except for state changes when resources are creating / updating / deleting
	// Where updates would be service side changes (updating customer infra, system maintenance, etc)
	// we will still do data updates here inorder to support imports, though they will only be partial as API's do
	// not return some sensitive data like credential information
	data.State = types.StringValue(describeResourceResponse.State)
	data.DataSourceType = types.StringValue(describeResourceResponse.DataSourceType)
	data.Description = types.StringPointerValue(describeResourceResponse.Description)
	data.SerialColumn = types.StringValue(describeResourceResponse.SerialColumn)
	data.PartitioningColumn = types.StringValue(describeResourceResponse.PartitioningColumn)

	data.DataSourceConfig, _ = types.MapValueFrom(ctx, types.StringType, describeResourceResponse.DataSourceConfig)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Ambar does not support resource updates, only credential rotations. Instead, all attributes
	// should include the PlanModifier indicating replacement is required on changes. RequiresReplace()
	var data dataSourceResourceModel
	var err error

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make the call to update the credentials
	var updateCredentialsRequest Ambar.UpdateResourceCredentialsRequest
	updateCredentialsRequest.ResourceId = data.ResourceId.ValueString()
	updateCredentialsRequest.Username = data.Username.ValueString()
	updateCredentialsRequest.Password = data.Password.ValueString()

	updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataSourceCredentials(ctx).UpdateResourceCredentialsRequest(updateCredentialsRequest).Execute()
	if err != nil || updateResourceResponse == nil || httpResponse == nil {
		resp.Diagnostics.AddError(
			"Error updating DataSource",
			"Could not update DataSource, unexpected error: "+err.Error(),
		)
		return
	}

	// partial state save in case of interrupt
	data.State = types.StringValue(updateResourceResponse.ResourceState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Wait for the update to complete
	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = data.ResourceId.ValueString()

	var describeResourceResponse *Ambar.DataSource

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, _, err = r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
		if err != nil {
			return
		}

		if describeResourceResponse.State == "READY" {
			break
		}
	}

	data.State = types.StringValue(describeResourceResponse.State)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	_, _, err := r.client.AmbarAPI.DeleteDataSource(ctx).DeleteResourceRequest(deleteDataSource).Execute()
	// Todo: Error handling as this call should not throw
	if err != nil {
		return
	}

	// Wait for confirmation the resource is Deleted via a ResourceNotFound error when describing it.
	var describeDataSource Ambar.DescribeResourceRequest
	describeDataSource.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		_, _, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeDataSource).Execute()
		if err != nil {
			return
		}
	}
}

func (r *dataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}
