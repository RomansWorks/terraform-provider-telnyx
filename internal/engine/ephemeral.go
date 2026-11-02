// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

type frameworkEphemeral struct {
	e *Ephemeral
	c *client.Client
}

var _ ephemeral.EphemeralResourceWithConfigure = &frameworkEphemeral{}

// NewEphemeral returns a Terraform ephemeral resource constructor for a
// descriptor.
func NewEphemeral(e *Ephemeral) func() ephemeral.EphemeralResource {
	return func() ephemeral.EphemeralResource { return &frameworkEphemeral{e: e} }
}

func (r *frameworkEphemeral) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = r.e.TypeName
}

func (r *frameworkEphemeral) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = r.e.Schema()
}

func (r *frameworkEphemeral) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

// Schema builds the ephemeral resource schema: path parameters are
// required, everything else computed and sensitive.
func (e *Ephemeral) Schema() eschema.Schema {
	attrs := make(map[string]eschema.Attribute, len(e.Attrs))
	for _, a := range e.Attrs {
		attrs[a.Name] = ephemeralAttr(a, len(a.PathParams) > 0)
	}
	return eschema.Schema{MarkdownDescription: e.Description, Attributes: attrs}
}

func ephemeralAttr(a *Attr, required bool) eschema.Attribute {
	desc := description(a)
	comp, sens := !required, !required || a.Sensitive
	switch a.Kind {
	case KindString, KindJSON:
		return eschema.StringAttribute{Required: required, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindInt:
		return eschema.Int64Attribute{Required: required, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindFloat:
		return eschema.Float64Attribute{Required: required, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindBool:
		return eschema.BoolAttribute{Required: required, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindList:
		if a.Elem != nil {
			return eschema.ListAttribute{ElementType: attrType(a.Elem), Computed: true, Sensitive: true, MarkdownDescription: desc}
		}
		return eschema.ListNestedAttribute{NestedObject: eschema.NestedAttributeObject{Attributes: ephemeralNested(a.Attrs)},
			Computed: true, Sensitive: true, MarkdownDescription: desc}
	case KindSet:
		if a.Elem != nil {
			return eschema.SetAttribute{ElementType: attrType(a.Elem), Computed: true, Sensitive: true, MarkdownDescription: desc}
		}
		return eschema.SetNestedAttribute{NestedObject: eschema.NestedAttributeObject{Attributes: ephemeralNested(a.Attrs)},
			Computed: true, Sensitive: true, MarkdownDescription: desc}
	case KindObject:
		return eschema.SingleNestedAttribute{Attributes: ephemeralNested(a.Attrs), Computed: true, Sensitive: true, MarkdownDescription: desc}
	}
	panic(fmt.Sprintf("engine: unknown kind for %s", a.Name))
}

// ephemeralNested builds nested attributes; their parent is sensitive.
func ephemeralNested(attrs []*Attr) map[string]eschema.Attribute {
	out := make(map[string]eschema.Attribute, len(attrs))
	for _, a := range attrs {
		out[a.Name] = ephemeralAttr(a, false)
	}
	return out
}

func (r *frameworkEphemeral) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	e := r.e
	config := fields(req.Config.Raw)
	p, err := fillPath(e.Attrs, e.Open.Path, config)
	if err == nil {
		var out []byte
		if out, err = r.c.Send(ctx, e.Open.Method, p, nil, "", nil); err == nil {
			var obj Object
			if obj, err = Decode(out, e.Open.Envelope); err == nil {
				resp.Result.Raw, err = objectFromJSON(e.Attrs, obj, config, modeData)
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to open "+e.TypeName, err.Error())
	}
}
