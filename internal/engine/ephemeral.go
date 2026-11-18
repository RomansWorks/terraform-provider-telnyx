// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		attrs[a.Name] = ephemeralAttr(a, a.input())
	}
	return eschema.Schema{MarkdownDescription: e.Description, Attributes: attrs}
}

// input reports whether an ephemeral attribute is an argument: a path
// parameter or a request body field. Everything else is a result.
func (a *Attr) input() bool {
	return len(a.PathParams) > 0 || a.InCreate
}

func ephemeralAttr(a *Attr, input bool) eschema.Attribute {
	desc := description(a)
	if input {
		// An argument is never compared with the API: no drift note.
		desc = description(&Attr{Description: a.Description, Enum: a.Enum, Kind: a.Kind, Required: a.Required, Optional: a.Optional, Local: true})
	}
	req := input && (a.Required || len(a.PathParams) > 0)
	opt := input && !req
	comp, sens := !input, !input || a.Sensitive
	switch a.Kind {
	case KindString, KindJSON:
		return eschema.StringAttribute{Required: req, Optional: opt, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindInt:
		return eschema.Int64Attribute{Required: req, Optional: opt, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindFloat:
		return eschema.Float64Attribute{Required: req, Optional: opt, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindBool:
		return eschema.BoolAttribute{Required: req, Optional: opt, Computed: comp, Sensitive: sens, MarkdownDescription: desc}
	case KindList:
		if input {
			return eschema.ListAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt, Sensitive: a.Sensitive, MarkdownDescription: desc}
		}
		if a.Elem != nil {
			return eschema.ListAttribute{ElementType: attrType(a.Elem), Computed: true, Sensitive: true, MarkdownDescription: desc}
		}
		return eschema.ListNestedAttribute{NestedObject: eschema.NestedAttributeObject{Attributes: ephemeralNested(a.Attrs)},
			Computed: true, Sensitive: true, MarkdownDescription: desc}
	case KindSet:
		if input {
			return eschema.SetAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt, Sensitive: a.Sensitive, MarkdownDescription: desc}
		}
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
		var out, data []byte
		contentType := ""
		var body Object
		if body, err = objectToJSON(e.Attrs, config, func(a *Attr) bool { return a.InCreate }); err == nil && hasInput(e.Attrs) {
			// Only the arguments the configuration sets are sent.
			data, err = json.Marshal(body)
			contentType = "application/json"
		}
		if err == nil {
			out, err = r.c.Send(ctx, e.Open.Method, p, nil, contentType, data)
		}
		if err == nil {
			var obj Object
			if e.Open.Text {
				obj = Object{"value": strings.TrimSpace(string(out))}
			} else {
				obj, err = Decode(out, e.Open.Envelope)
			}
			if err == nil {
				resp.Result.Raw, err = objectFromJSON(e.Attrs, obj, config, modeData)
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to open "+e.TypeName, err.Error())
	}
}

// hasInput reports whether the operation takes a request body.
func hasInput(attrs []*Attr) bool {
	for _, a := range attrs {
		if a.InCreate && !a.Local {
			return true
		}
	}
	return false
}
