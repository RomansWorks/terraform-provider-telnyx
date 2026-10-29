// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package engine

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

type frameworkDataSource struct {
	d *Resource
	c *client.Client
}

var _ datasource.DataSourceWithConfigure = &frameworkDataSource{}

// NewDataSource returns a Terraform data source constructor for a
// descriptor. The data source looks the object up by the import attributes.
func NewDataSource(d *Resource) func() datasource.DataSource {
	return func() datasource.DataSource { return &frameworkDataSource{d: d} }
}

func (r *frameworkDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = r.d.TypeName
}

func (r *frameworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = r.d.DataSourceSchema()
}

func (r *frameworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.c = c
}

func (r *frameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := fields(req.Config.Raw)
	obj, err := Get(ctx, r.c, r.d, config)
	if err == nil && r.d.Hooks.AfterRead != nil {
		obj, err = r.d.Hooks.AfterRead(ctx, obj)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+r.d.TypeName, err.Error())
		return
	}
	v, err := objectFromJSON(r.d.dataAttrs(), obj, config, modeData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+r.d.TypeName, err.Error())
		return
	}
	resp.State.Raw = v
}
