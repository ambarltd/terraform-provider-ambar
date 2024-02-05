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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DataDestinationResource{}
var _ resource.ResourceWithImportState = &DataDestinationResource{}

func NewDataDestinationResource() resource.Resource {
	return &DataDestinationResource{}
}

// DataDestinationResource defines the resource implementation.
type DataDestinationResource struct {
	client *Ambar.APIClient
}

// DataDestinationResourceModel describes the resource data model.
type dataDestinationResourceModel struct {
	FilterIds           types.List   `tfsdk:"filter_ids"`
	Description         types.String `tfsdk:"description"`
	DestinationEndpoint types.String `tfsdk:"destination_endpoint"`
	Username            types.String `tfsdk:"username"`
	Password            types.String `tfsdk:"password"`
	State               types.String `tfsdk:"state"`
	ResourceId          types.String `tfsdk:"resource_id"`
}

func (r *DataDestinationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_destination"
}

func (r *DataDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ambar DataDestination resource. Represents details about a Destination HTTP server you have configured to receive filtered record sequences from Ambar.",
		Description:         "Ambar DataDestination resource. Represents details about a Destination HTTP server you have configured to receive filtered record sequences from Ambar.",

		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A List of Ambar resource ids belonging to Ambar Filter resources which should be used with this DataDestination. These control what DataSources and applied filters will be delivered to your destination. Note that a DataSource can only be used once per DataDestination.",
				Description:         "A List of Ambar resource ids belonging to Ambar Filter resources which should be used with this DataDestination. These control what DataSources and applied filters will be delivered to your destination. Note that a DataSource can only be used once per DataDestination.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A user friendly description of this DataDestination. Use the description filed to help augment information about this DataDestination which may may not be apparent from describing the resource, such as details about the filtered record sequences being sent.",
				Description:         "A user friendly description of this DataDestination. Use the description filed to help augment information about this DataDestination which may may not be apparent from describing the resource, such as details about the filtered record sequences being sent.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_endpoint": schema.StringAttribute{
				MarkdownDescription: "The HTTP endpoint where Ambar will send your filtered record sequences to.",
				Description:         "The HTTP endpoint where Ambar will send your filtered record sequences to.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "A username credential which Ambar can use to communicate with your destination.",
				Description:         "A username credential which Ambar can use to communicate with your destination.",
				Required:            true,
				Sensitive:           true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "A password credential which Ambar can use to communicate with your destination.",
				Description:         "A password credential which Ambar can use to communicate with your destination.",
				Required:            true,
				Sensitive:           true,
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

func (r *DataDestinationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DataDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan dataDestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createDataDestination Ambar.CreateDataDestinationRequest

	elements := make([]string, 0, len(plan.FilterIds.Elements()))
	_ = plan.FilterIds.ElementsAs(ctx, &elements, false)

	createDataDestination.FilterIds = elements

	createDataDestination.Description = plan.Description.ValueStringPointer()
	createDataDestination.Username = plan.Username.ValueString()
	createDataDestination.Password = plan.Password.ValueString()
	createDataDestination.DestinationEndpoint = plan.DestinationEndpoint.ValueString()

	// Create the API call and execute it
	createResourceResponse, httpResponse, err := r.client.AmbarAPI.CreateDataDestination(ctx).CreateDataDestinationRequest(createDataDestination).Execute()
	if err != nil || createResourceResponse == nil || httpResponse == nil {
		resp.Diagnostics.AddError(
			"Error creating DataDestination",
			"Could not create DataDestination, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ResourceId = types.StringValue(createResourceResponse.ResourceId)
	plan.State = types.StringValue(createResourceResponse.State)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = createResourceResponse.ResourceId

	var describeResourceResponse *Ambar.DataDestination

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, _, err = r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()
		if err != nil {
			return
		}

		if describeResourceResponse.State == "READY" {
			break
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan.ResourceId = types.StringValue(createResourceResponse.ResourceId)
	plan.State = types.StringValue(describeResourceResponse.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DataDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataDestinationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the latest state from the Ambar describe API
	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = data.ResourceId.ValueString()

	describeResourceResponse, _, err := r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()
	// Todo: Handle ResourceNotFoundException gracefully per https://developer.hashicorp.com/terraform/plugin/framework/resources/read#recommendations
	if err != nil {
		return
	}

	data.State = types.StringValue(describeResourceResponse.State)
	data.DestinationEndpoint = types.StringValue(describeResourceResponse.DestinationEndpoint)
	data.Description = types.StringPointerValue(describeResourceResponse.Description)

	data.FilterIds, _ = types.ListValueFrom(ctx, types.StringType, describeResourceResponse.FilterIds)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Ambar does not support resource updates, only credential rotations. Instead, all attributes
	// should include the PlanModifier indicating replacement is required on changes. RequiresReplace()
	var data dataDestinationResourceModel
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

	updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataDestinationCredentials(ctx).UpdateResourceCredentialsRequest(updateCredentialsRequest).Execute()
	if err != nil || updateResourceResponse == nil || httpResponse == nil {
		resp.Diagnostics.AddError(
			"Error updating DataDestination",
			"Could not update DataDestination, unexpected error: "+err.Error(),
		)
		return
	}

	// partial state save in case of interrupt
	data.State = types.StringValue(updateResourceResponse.State)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Wait for the update to complete
	var describeResourceResponse *Ambar.DataDestination
	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, _, err = r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()
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

func (r *DataDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataDestinationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the DataDestination
	var deleteDataDestination Ambar.DeleteResourceRequest
	deleteDataDestination.ResourceId = data.ResourceId.ValueString()

	_, _, err := r.client.AmbarAPI.DeleteDataDestination(ctx).DeleteResourceRequest(deleteDataDestination).Execute()
	// Todo: Error handling as this call should not throw
	if err != nil {
		return
	}

	// Wait for confirmation the resource is Deleted via a ResourceNotFound error when describing it.
	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		_, _, err := r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()
		if err != nil {
			return
		}
	}
}

func (r *DataDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}
