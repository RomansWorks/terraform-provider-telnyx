// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// maxItemsAttr caps how many items a list data source reads.
const maxItemsAttr = "max_items"

type frameworkList struct {
	l *List
	c *client.Client
}

var _ datasource.DataSourceWithConfigure = &frameworkList{}

// NewListDataSource returns a Terraform data source constructor for a list
// descriptor.
func NewListDataSource(l *List) func() datasource.DataSource {
	return func() datasource.DataSource { return &frameworkList{l: l} }
}

func (r *frameworkList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = r.l.TypeName
}

func (r *frameworkList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = r.l.Schema()
}

func (r *frameworkList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema builds the list data source schema: path parameters are required,
// query parameters and max_items optional, and the items computed.
func (l *List) Schema() dschema.Schema {
	attrs := map[string]dschema.Attribute{
		maxItemsAttr: dschema.Int64Attribute{Optional: true,
			MarkdownDescription: "Stop reading after this many items. By default every page is read."},
		l.Items.Name: dataAttr(l.Items, false),
	}
	for _, a := range l.Params {
		// Parameters are sent, never read back: no drift note.
		p := *a
		p.InResponse = true
		desc := description(&p)
		required := len(a.PathParams) > 0
		switch a.Kind {
		case KindInt:
			attrs[a.Name] = dschema.Int64Attribute{Required: required, Optional: !required, MarkdownDescription: desc}
		case KindFloat:
			attrs[a.Name] = dschema.Float64Attribute{Required: required, Optional: !required, MarkdownDescription: desc}
		case KindBool:
			attrs[a.Name] = dschema.BoolAttribute{Required: required, Optional: !required, MarkdownDescription: desc}
		case KindList:
			attrs[a.Name] = dschema.ListAttribute{ElementType: attrType(a.Elem), Optional: true, MarkdownDescription: desc}
		default:
			attrs[a.Name] = dschema.StringAttribute{Required: required, Optional: !required, MarkdownDescription: desc}
		}
	}
	return dschema.Schema{MarkdownDescription: l.Description, Attributes: attrs}
}

// objectType is the state type of the list data source.
func (l *List) objectType() tftypes.Object {
	types := map[string]tftypes.Type{maxItemsAttr: tftypes.Number, l.Items.Name: l.Items.TerraformType()}
	for _, a := range l.Params {
		types[a.Name] = a.TerraformType()
	}
	return tftypes.Object{AttributeTypes: types}
}

func (r *frameworkList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	l := r.l
	config := fields(req.Config.Raw)
	p, err := fillPath(l.Params, l.Op.Path, config)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+l.TypeName, err.Error())
		return
	}
	query, err := queryValues(l.Params, config)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+l.TypeName, err.Error())
		return
	}
	limit := 0
	if v := config[maxItemsAttr]; v.IsKnown() && !v.IsNull() {
		var n big.Float
		if err := v.As(&n); err == nil {
			i, _ := n.Int64()
			limit = int(i)
		}
	}

	var items []any
	err = Each(ctx, r.c, l.Op, p, query, func(it Object) bool {
		if limit > 0 && len(items) >= limit {
			return false
		}
		if l.Items.Elem != nil && l.Items.Elem.Kind != KindJSON {
			// Plain values come wrapped by decodeList.
			items = append(items, it["value"])
		} else {
			items = append(items, it)
		}
		return limit <= 0 || len(items) < limit
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+l.TypeName, err.Error())
		return
	}
	if items == nil {
		items = []any{}
	}
	v, err := fromJSON(l.Items, items, tftypes.NewValue(l.Items.TerraformType(), nil), modeData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+l.TypeName, err.Error())
		return
	}
	typ := l.objectType()
	out := map[string]tftypes.Value{l.Items.Name: v}
	for name, t := range typ.AttributeTypes {
		if name == l.Items.Name {
			continue
		}
		if c, ok := config[name]; ok {
			out[name] = c
		} else {
			out[name] = tftypes.NewValue(t, nil)
		}
	}
	resp.State.Raw = tftypes.NewValue(typ, out)
}

// queryValues encodes the set query parameters. Lists are joined with
// commas, which is how Telnyx filters take several values.
func queryValues(params []*Attr, config Values) (url.Values, error) {
	q := url.Values{}
	for _, a := range params {
		if len(a.PathParams) > 0 {
			continue
		}
		v := config[a.Name]
		if !v.IsKnown() || v.IsNull() {
			continue
		}
		raw, ok, err := toJSON(a, v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}
		if !ok {
			continue
		}
		if xs, isList := raw.([]any); isList {
			parts := make([]string, len(xs))
			for i, x := range xs {
				parts[i] = scalarString(x)
			}
			q.Set(a.JSON, strings.Join(parts, ","))
			continue
		}
		q.Set(a.JSON, scalarString(raw))
	}
	return q, nil
}
