// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &QuanianitisMailResource{}
var _ resource.ResourceWithImportState = &QuanianitisMailResource{}

func NewQuanianitisMailResource() resource.Resource {
	return &QuanianitisMailResource{}
}

// QuanianitisMailResource defines the resource implementation.
type QuanianitisMailResource struct {
	client *Client
}

// QuanianitisMailResourceModel describes the resource data model.
type QuanianitisMailResourceModel struct {
	Id      types.String `tfsdk:"id"`
	From    types.String `tfsdk:"from"`
	To      types.String `tfsdk:"to"`
	Subject types.String `tfsdk:"subject"`
	Content types.String `tfsdk:"content"`
}

func (r *QuanianitisMailResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mail"
}

func (r *QuanianitisMailResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Outbound Mail Forwarding for quanianitis.com domain",

		Attributes: map[string]schema.Attribute{
			"from": schema.StringAttribute{
				MarkdownDescription: "work@quanianitis.com",
				Required:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "The email address of the recipient",
				Required:            true,
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "The subject of the email",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The plain text content of the email",
				Required:            true,
			},
		},
	}
}

func (r *QuanianitisMailResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *QuanianitisMailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data QuanianitisMailResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(time.Now().Format(time.RFC3339))

	mail := map[string]string{
		"from_address":       data.From.ValueString(),
		"to_address":         data.To.ValueString(),
		"subject":            data.Subject.ValueString(),
		"plain_text_content": data.Content.ValueString(),
	}
	mailBytes, err := json.Marshal(mail)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling and transforming submitted data to a Mail-compatible format",
			fmt.Sprintf("Unable to marshal submitted data to Mail-compatbile format", err),
		)
		return
	}

	functionURL := r.client.Endpoint
	http, err := http.NewRequest("POST", functionURL, bytes.NewBuffer(mailBytes))
	if err != nil {
		resp.Diagnostics.HasError()
		return
	}

	http.Header.Set("Content-Type", "application/json")

	httpResponse, err := r.client.Client.Do(http)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error forwarding mail",
			fmt.Sprintf("Unable to forward mail: %s", err),
		)
		return
	}
	defer http.Body.Close()

	if httpResponse.StatusCode != httpResponse.StatusCode {
		resp.Diagnostics.AddError(
			"Error response from Cloud Function",
			fmt.Sprintf("Error returned status code %d", httpResponse.StatusCode),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QuanianitisMailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QuanianitisMailResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QuanianitisMailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data QuanianitisMailResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QuanianitisMailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data QuanianitisMailResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *QuanianitisMailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
