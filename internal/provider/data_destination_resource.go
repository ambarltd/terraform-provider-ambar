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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strconv"
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
				Required:            true,
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
		tflog.Debug(ctx, "StatusCode: "+httpResponse.Status)
		httpBody, _ := io.ReadAll(httpResponse.Body)
		errString := string(httpBody)
		tflog.Debug(ctx, errString)
		resp.Diagnostics.AddError(
			"Error creating DataDestination",
			"Could not create DataDestination: "+AmbarApiErrorToTerraformErrorString(errString),
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
			tflog.Debug(ctx, "Got error while waiting for resource to become ready: "+err.Error())
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

	describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()
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

	data.State = types.StringValue(describeResourceResponse.State)
	data.DestinationEndpoint = types.StringValue(describeResourceResponse.DestinationEndpoint)
	data.Description = types.StringPointerValue(describeResourceResponse.Description)

	data.FilterIds, _ = types.ListValueFrom(ctx, types.StringType, describeResourceResponse.FilterIds)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Ambar supports resource updates for credential rotations, and destinationEndpoints. Instead, all attributes
	// should include the PlanModifier indicating replacement is required on changes. RequiresReplace()
	var plan dataDestinationResourceModel
	var current dataDestinationResourceModel
	var err error

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &current)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// We need to validate we can perform the change in a single operation
	if (plan.Username.ValueString() != current.Username.ValueString() || plan.Password.ValueString() != current.Password.ValueString()) && plan.DestinationEndpoint.ValueString() != current.DestinationEndpoint.ValueString() {
		resp.Diagnostics.AddError("Invalid parameter change combination.",
			"When updating Ambar resources, perform credential changes independent of any other updates on resources.")
		return
	}

	var updateResourceResponse Ambar.ResourceStateChangeResponse

	if plan.Username.ValueString() != current.Username.ValueString() || plan.Password.ValueString() != current.Password.ValueString() {
		// Make the call to update the credentials if that is what is requested
		var updateCredentialsRequest Ambar.UpdateResourceCredentialsRequest
		updateCredentialsRequest.ResourceId = plan.ResourceId.ValueString()
		updateCredentialsRequest.Username = plan.Username.ValueString()
		updateCredentialsRequest.Password = plan.Password.ValueString()

		updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataDestinationCredentials(ctx).UpdateResourceCredentialsRequest(updateCredentialsRequest).Execute()
		if err != nil || updateResourceResponse == nil || httpResponse == nil {
			resp.Diagnostics.AddError(
				"Error updating DataDestination",
				"Could not update DataDestination, unexpected error: "+err.Error(),
			)
			return
		}

	} else {
		// Make the call to update the endpoint.
		var updateDestinationRequest Ambar.UpdateDataDestinationRequest
		updateDestinationRequest.ResourceId = plan.ResourceId.ValueString()
		updateDestinationRequest.DestinationEndpoint = plan.DestinationEndpoint.ValueString()

		updateResourceResponse, httpResponse, err := r.client.AmbarAPI.UpdateDataDestination(ctx).UpdateDataDestinationRequest(updateDestinationRequest).Execute()
		if err != nil || updateResourceResponse == nil || httpResponse == nil {
			resp.Diagnostics.AddError(
				"Error updating DataDestination",
				"Could not update DataDestination, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// partial state save in case of interrupt
	plan.State = types.StringValue(updateResourceResponse.State)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// Wait for the update to complete
	var describeResourceResponse *Ambar.DataDestination
	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = plan.ResourceId.ValueString()

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

	plan.State = types.StringValue(describeResourceResponse.State)

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	deleteResponse, httpResponse, err := r.client.AmbarAPI.DeleteDataDestination(ctx).DeleteResourceRequest(deleteDataDestination).Execute()
	if err != nil {
		tflog.Error(ctx, "Got error: "+err.Error())

		if httpResponse.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "Resource was not found. Removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Error(ctx, "Unexpected error, dumping logs and returning.")
		tflog.Error(ctx, "Got http response, code: "+strconv.Itoa(httpResponse.StatusCode)+", status: "+httpResponse.Status)
		resp.Diagnostics.AddError("Unable to delete DataDestination resource.", err.Error())
		return
	}
	tflog.Info(ctx, "Got deleteResponse: "+deleteResponse.State)

	var describeDataDestination Ambar.DescribeResourceRequest
	describeDataDestination.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeDataDestination(ctx).DescribeResourceRequest(describeDataDestination).Execute()

		if err != nil {
			tflog.Error(ctx, "Got error: "+err.Error())

			if httpResponse.StatusCode == http.StatusNotFound {
				tflog.Info(ctx, "Resource was not found. This is expected for delete, returning.")
				return
			}

			tflog.Error(ctx, "Unexpected error, dumping logs and returning.")
			tflog.Error(ctx, "Got http response, code: "+strconv.Itoa(httpResponse.StatusCode)+", status: "+httpResponse.Status)
			resp.Diagnostics.AddError("Unable to read DataDestination resource to confirm deletion. Got error.", err.Error())
			return
		}

		tflog.Debug(ctx, "Waiting for resource to complete deletion. Current state: "+describeResourceResponse.State)
	}
}

func (r *DataDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}
