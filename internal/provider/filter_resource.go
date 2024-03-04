// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	Ambar "github.com/ambarltd/ambar_go_client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
var _ resource.Resource = &FilterResource{}
var _ resource.ResourceWithImportState = &FilterResource{}

func NewFilterResource() resource.Resource {
	return &FilterResource{}
}

// FilterResource defines the resource implementation.
type FilterResource struct {
	client *Ambar.APIClient
}

// FilterResourceModel describes the resource data model.
type filterResourceModel struct {
	DataSourceId   types.String `tfsdk:"data_source_id"`
	Description    types.String `tfsdk:"description"`
	FilterContents types.String `tfsdk:"filter_contents"`
	State          types.String `tfsdk:"state"`
	ResourceId     types.String `tfsdk:"resource_id"`
}

func (r *FilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

func (r *FilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ambar Filter resource. Represents details about an Ambar DataSource to be read from, and filtering to be done on its record sequences before delivery.",
		Description:         "Ambar Filter resource. Represents details about an Ambar DataSource to be read from, and filtering to be done on its record sequences before delivery.",

		Attributes: map[string]schema.Attribute{
			"data_source_id": schema.StringAttribute{
				MarkdownDescription: "An Ambar resource id belonging to an Ambar DataSource for which this Filter should be applied to.",
				Description:         "An Ambar resource id belonging to an Ambar DataSource for which this Filter should be applied to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A user friendly description of this Filter. Use the description field to help augment information about this Filter which may not be apparent from describing the resource, such as what it is filtering.",
				Description:         "A user friendly description of this Filter. Use the description field to help augment information about this Filter which may not be apparent from describing the resource, such as what it is filtering.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filter_contents": schema.StringAttribute{
				MarkdownDescription: "A string filter statement using Ambar Filter syntax. See [Ambar documentation](https://docs.ambar.cloud) for more details on valid Ambar filtering operations on record sequences.",
				Description:         "A string filter statement using Ambar Filter syntax. See [Ambar documentation](https://docs.ambar.cloud) for more details on valid Ambar filtering operations on record sequences.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *FilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan filterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createFilter Ambar.CreateFilterRequest
	createFilter.Description = plan.Description.ValueStringPointer()

	// Filters we encode the content string in base64 for API requests to compact complex filters and prevent white spaces
	// from giving any issues. If the customer already encoded the contents, great. But it is not required. So we will
	// check if they already did it, and otherwise encode it.
	_, err := base64.StdEncoding.DecodeString(plan.FilterContents.ValueString())
	if err != nil {
		// decoding failed, and we will want to encode and assign
		encodedContents := base64.StdEncoding.EncodeToString([]byte(plan.FilterContents.ValueString()))
		createFilter.FilterContents = encodedContents
	} else {
		// customer already base64 encoded, so just pass it on through.
		createFilter.FilterContents = plan.FilterContents.ValueString()
	}

	createFilter.DataSourceId = plan.DataSourceId.ValueString()

	// Create the API call and execute it
	createResourceResponse, httpResponse, err := r.client.AmbarAPI.CreateFilter(ctx).CreateFilterRequest(createFilter).Execute()
	if err != nil || createResourceResponse == nil || httpResponse == nil {
		tflog.Debug(ctx, "StatusCode: "+httpResponse.Status)
		httpBody, _ := io.ReadAll(httpResponse.Body)
		errString := string(httpBody)
		tflog.Debug(ctx, errString)
		resp.Diagnostics.AddError(
			"Error creating Filter",
			"Could not create Filter: "+AmbarApiErrorToTerraformErrorString(errString),
		)
		return
	}

	// Give a few seconds for eventual consistency / resource to finish creating.
	time.Sleep(5 * time.Second)

	var describeFilter Ambar.DescribeResourceRequest
	describeFilter.ResourceId = createResourceResponse.ResourceId

	describeResourceResponse, _, err := r.client.AmbarAPI.DescribeFilter(ctx).DescribeResourceRequest(describeFilter).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error while describing Filter resource", err.Error())
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ResourceId = types.StringValue(createResourceResponse.ResourceId)
	plan.State = types.StringValue(describeResourceResponse.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *FilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data filterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the latest state from the Ambar describe API
	var describeFilter Ambar.DescribeResourceRequest
	describeFilter.ResourceId = data.ResourceId.ValueString()

	describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeFilter(ctx).DescribeResourceRequest(describeFilter).Execute()
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
	data.Description = types.StringPointerValue(describeResourceResponse.Description)
	data.DataSourceId = types.StringValue(describeResourceResponse.DataSourceId)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Ambar does not support resource updates, so there is nothing to do in this method. Instead, all attributes
	// should include the PlanModifier indicating replacement is required on changes. RequiresReplace()

	var data filterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data filterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Filter
	var deleteFilter Ambar.DeleteResourceRequest
	deleteFilter.ResourceId = data.ResourceId.ValueString()

	deleteResponse, httpResponse, err := r.client.AmbarAPI.DeleteFilter(ctx).DeleteResourceRequest(deleteFilter).Execute()
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

	var describeFilter Ambar.DescribeResourceRequest
	describeFilter.ResourceId = data.ResourceId.ValueString()

	for {
		time.Sleep(10 * time.Second)

		describeResourceResponse, httpResponse, err := r.client.AmbarAPI.DescribeDataSource(ctx).DescribeResourceRequest(describeFilter).Execute()

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

func (r *FilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}
