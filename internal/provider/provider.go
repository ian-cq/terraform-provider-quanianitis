// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

	// "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	// "github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// resources "github.com/ian-chan-ml/terraform-provider-quanianitis/internal/resources"
	// services "github.com/ian-chan-ml/terraform-provider-quanianitis/internal/services"
)

// Ensure quanianitisProvider satisfies various provider interfaces.
var _ provider.Provider = &quanianitisProvider{}

// var _ provider.ProviderWithFunctions = &quanianitisProvider{}

// quanianitisProvider defines the provider implementation.
type quanianitisProvider struct {
	version string
}

type Client struct {
	Endpoint string
	Client   *http.Client
}

// quanianitisProviderModel describes the provider data model.
type quanianitisProviderModel struct {
	Endpoint              types.String `tfsdk:"endpoint"`
	Gcloud_identity_token types.String `tfsdk:"gcloud_identity_token"`
}

func (p *quanianitisProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "quanianitis"
	resp.Version = p.version
}

func (p *quanianitisProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint for cloud function mail forwarding for quanianitis.com domain",
				Optional:            true,
				Required:            false,
				Sensitive:           true,
			},
			"gcloud_identity_token": schema.StringAttribute{
				MarkdownDescription: "Authenticate and authorize to send and invoke mail forward with this function",
				Optional:            true,
				Required:            false,
				Sensitive:           true,
			},
		},
	}
}

func (p *quanianitisProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config quanianitisProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Summary (Endpoint)",
			"Detail (Endpoint)",
		)
	}

	if config.Gcloud_identity_token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("gcloud_identity_token"),
			"Summary (Gcloud_identity_token)",
			"Detail (Gcloud_identity_token)",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := "https://asia-southeast1-personal-quanianitis.cloudfunctions.net/sendgrid"
	gcloud_identity_token := os.Getenv("GCLOUD_IDENTITY_TOKEN")

	if !config.Endpoint.IsNull() {
		config.Endpoint = types.StringValue(endpoint)
	}

	if !config.Gcloud_identity_token.IsNull() {
		config.Gcloud_identity_token = types.StringValue(gcloud_identity_token)
	}

	if gcloud_identity_token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("gcloud_identity_token"),
			"Summary (Gcloud_identity_token)",
			"Detail (Gcloud_identity_token)",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := &http.Client{
		Transport: &transportWithAuth{
			token: gcloud_identity_token,
			base:  http.DefaultTransport,
		},
	}
	resp.DataSourceData = client
	resp.ResourceData = &Client{
		Endpoint: config.Endpoint.ValueString(),
		Client:   client,
	}
}

func (p *quanianitisProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewQuanianitisMailResource,
	}
}

// func (p *quanianitisProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
// 	return []func() datasource.DataSource{
// 		NewExampleDataSource,
// 	}
// }

// func (p *quanianitisProvider) Functions(ctx context.Context) []func() function.Function {
// 	return []func() function.Function{
// 		NewExampleFunction,
// 	}
// }

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &quanianitisProvider{
			version: version,
		}
	}
}

type transportWithAuth struct {
	token string
	base  http.RoundTripper
}

func (t *transportWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}
