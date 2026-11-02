// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	aschema "github.com/hashicorp/terraform-plugin-framework/action/schema"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

type frameworkAction struct {
	a *Action
	c *client.Client
}

var _ action.ActionWithConfigure = &frameworkAction{}

// NewAction returns a Terraform action constructor for a descriptor.
func NewAction(a *Action) func() action.Action {
	return func() action.Action { return &frameworkAction{a: a} }
}

func (r *frameworkAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = r.a.TypeName
}

func (r *frameworkAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = r.a.Schema()
}

func (r *frameworkAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

// Schema builds the action schema: every attribute is configuration.
func (a *Action) Schema() aschema.Schema {
	attrs := make(map[string]aschema.Attribute, len(a.Attrs))
	for _, at := range a.Attrs {
		attrs[at.Name] = actionAttr(at)
	}
	return aschema.Schema{MarkdownDescription: a.Description, Attributes: attrs}
}

func actionAttr(a *Attr) aschema.Attribute {
	req, opt := a.Required, !a.Required
	desc := description(&Attr{Description: a.Description, Enum: a.Enum, Kind: a.Kind, Required: req, Optional: opt, Local: true})
	switch a.Kind {
	case KindString, KindJSON:
		return aschema.StringAttribute{Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindInt:
		return aschema.Int64Attribute{Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindFloat:
		return aschema.Float64Attribute{Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindBool:
		return aschema.BoolAttribute{Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindList:
		if a.Elem != nil {
			return aschema.ListAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt,
				DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
		}
		return aschema.ListNestedAttribute{NestedObject: aschema.NestedAttributeObject{Attributes: actionNested(a.Attrs)},
			Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindSet:
		if a.Elem != nil {
			return aschema.SetAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt,
				DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
		}
		return aschema.SetNestedAttribute{NestedObject: aschema.NestedAttributeObject{Attributes: actionNested(a.Attrs)},
			Required: req, Optional: opt, DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	case KindObject:
		return aschema.SingleNestedAttribute{Attributes: actionNested(a.Attrs), Required: req, Optional: opt,
			DeprecationMessage: a.Deprecated, MarkdownDescription: desc}
	}
	panic(fmt.Sprintf("engine: unknown kind for %s", a.Name))
}

func actionNested(attrs []*Attr) map[string]aschema.Attribute {
	out := make(map[string]aschema.Attribute, len(attrs))
	for _, a := range attrs {
		out[a.Name] = actionAttr(a)
	}
	return out
}

func (r *frameworkAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	a := r.a
	config := fields(req.Config.Raw)
	p, err := fillPath(a.Attrs, a.Op.Path, config)
	if err != nil {
		resp.Diagnostics.AddError("Unable to run "+a.TypeName, err.Error())
		return
	}
	body, err := objectToJSON(a.Attrs, config, func(at *Attr) bool { return !at.Local })
	if err != nil {
		resp.Diagnostics.AddError("Unable to run "+a.TypeName, err.Error())
		return
	}
	var data []byte
	contentType := ""
	if hasBody(a.Attrs) {
		if data, err = json.Marshal(body); err != nil {
			resp.Diagnostics.AddError("Unable to run "+a.TypeName, err.Error())
			return
		}
		contentType = "application/json"
	}
	if _, err := r.c.Send(ctx, a.Op.Method, p, nil, contentType, data); err != nil {
		resp.Diagnostics.AddError("Unable to run "+a.TypeName, err.Error())
		return
	}
	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("%s: %s %s done", a.TypeName, a.Op.Method, p)})
	}
}

// hasBody reports whether the operation takes a request body.
func hasBody(attrs []*Attr) bool {
	for _, a := range attrs {
		if !a.Local {
			return true
		}
	}
	return false
}
