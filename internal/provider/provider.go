// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

var _ provider.Provider = &TelnyxProvider{}

// TelnyxProvider is the Telnyx Terraform provider.
type TelnyxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TelnyxProviderModel describes the provider configuration block.
type TelnyxProviderModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	BaseURL    types.String `tfsdk:"base_url"`
	MaxRetries types.Int64  `tfsdk:"max_retries"`
}

func (p *TelnyxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "telnyx"
	resp.Version = p.version
}

func (p *TelnyxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage [Telnyx](https://telnyx.com) resources: numbers, voice, messaging, wireless, networking, storage and AI assistants.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Telnyx API v2 key. Can also be set with the `TELNYX_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Telnyx API. Defaults to `" + client.DefaultBaseURL + "`. Can also be set with `TELNYX_BASE_URL`.",
				Optional:            true,
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "How many times a rate-limited or failed idempotent request is retried. Defaults to `4`. Can also be set with `TELNYX_MAX_RETRIES`.",
				Optional:            true,
			},
		},
	}
}

func (p *TelnyxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TelnyxProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Unknown Telnyx API key",
			"The API key is not known until apply. Set it to a known value or use the TELNYX_API_KEY environment variable.")
		return
	}

	apiKey := stringOrEnv(data.APIKey, "TELNYX_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Missing Telnyx API key",
			"Set api_key in the provider block or the TELNYX_API_KEY environment variable.")
		return
	}

	maxRetries := client.DefaultMaxRetries
	switch {
	case !data.MaxRetries.IsNull() && !data.MaxRetries.IsUnknown():
		maxRetries = int(data.MaxRetries.ValueInt64())
	case os.Getenv("TELNYX_MAX_RETRIES") != "":
		n, err := strconv.Atoi(os.Getenv("TELNYX_MAX_RETRIES"))
		if err != nil {
			resp.Diagnostics.AddError("Invalid TELNYX_MAX_RETRIES", "TELNYX_MAX_RETRIES must be a whole number.")
			return
		}
		maxRetries = n
	}

	c, err := client.New(client.Config{
		APIKey:     apiKey,
		BaseURL:    stringOrEnv(data.BaseURL, "TELNYX_BASE_URL"),
		MaxRetries: maxRetries,
		UserAgent:  "terraform-provider-telnyx/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure the Telnyx client", err.Error())
		return
	}
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *TelnyxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *TelnyxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// New returns a constructor for the provider at the given version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TelnyxProvider{
			version: version,
		}
	}
}

func stringOrEnv(v types.String, env string) string {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v.ValueString()
	}
	return os.Getenv(env)
}
