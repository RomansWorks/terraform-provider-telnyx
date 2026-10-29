// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// DefaultCreateTimeout bounds the Wait hook when no timeout is configured.
const DefaultCreateTimeout = 10 * time.Minute

type frameworkResource struct {
	d *Resource
	c *client.Client
}

var (
	_ resource.ResourceWithConfigure        = &frameworkResource{}
	_ resource.ResourceWithImportState      = &frameworkResource{}
	_ resource.ResourceWithModifyPlan       = &frameworkResource{}
	_ resource.ResourceWithConfigValidators = &frameworkResource{}
)

// NewResource returns a Terraform resource constructor for a descriptor.
func NewResource(d *Resource) func() resource.Resource {
	return func() resource.Resource { return &frameworkResource{d: d} }
}

func (r *frameworkResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.d.TypeName
}

func (r *frameworkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.d.ResourceSchema(ctx)
}

func (r *frameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *frameworkResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	var out []resource.ConfigValidator
	for _, group := range r.d.ExactlyOne {
		exprs := make([]path.Expression, len(group))
		for i, n := range group {
			exprs[i] = path.MatchRoot(n)
		}
		out = append(out, resourcevalidator.ExactlyOneOf(exprs...))
	}
	return out
}

func (r *frameworkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		if r.d.Delete == nil && !req.State.Raw.IsNull() {
			resp.Diagnostics.AddWarning("Object stays in Telnyx", r.deleteWarning(fields(req.State.Raw)))
		}
		return
	}
	if r.d.Hooks.ModifyPlan == nil {
		return
	}
	plan := fields(req.Plan.Raw)
	var state Values
	if !req.State.Raw.IsNull() {
		state = fields(req.State.Raw)
	}
	replace, diags := r.d.Hooks.ModifyPlan(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Plan.Raw = tftypes.NewValue(req.Plan.Raw.Type(), map[string]tftypes.Value(plan))
	for _, n := range replace {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root(n))
	}
}

func (r *frameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := fields(req.Plan.Raw)
	d := r.d

	var obj Object
	var err error
	if d.Hooks.Create != nil {
		obj, err = d.Hooks.Create(ctx, r.c, d, plan)
	} else {
		obj, err = r.send(ctx, d.Create, plan, "create", func(a *Attr) bool { return a.InCreate }, plan)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create "+d.TypeName, err.Error())
		return
	}

	// Fill the id (and other response fields) so read and wait can find it.
	partial, err := r.toState(ctx, obj, plan, modeApply, req.Plan.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.TypeName+" create response", err.Error())
		return
	}
	vals := fields(partial)
	// Save the id straight away so a failed wait or read leaves the object
	// tracked instead of orphaned.
	resp.State.Raw = partial

	if d.Hooks.Wait != nil {
		var t timeouts.Value
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(timeoutsAttr), &t)...)
		timeout, diags := t.Create(ctx, DefaultCreateTimeout)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		wctx, cancel := context.WithTimeout(ctx, timeout)
		obj, err = d.Hooks.Wait(wctx, r.c, d, obj)
		cancel()
		if err != nil {
			resp.Diagnostics.AddError(d.TypeName+" did not become ready", err.Error())
			return
		}
	}

	// Attributes only an update accepts are applied straight after create.
	// A PUT replaces the whole object, so it cannot carry them alone.
	if d.Update != nil && d.Update.Method != http.MethodPut {
		late := func(a *Attr) bool {
			v := plan[a.Name]
			return a.InUpdate && !a.InCreate && !a.Local && v.IsKnown() && !v.IsNull()
		}
		if body, err := objectToJSON(d.Attrs, plan, late); err == nil && len(body) > 0 {
			if _, err := r.send(ctx, d.Update, plan, "update", late, vals); err != nil {
				resp.Diagnostics.AddError("Unable to update "+d.TypeName+" after create", err.Error())
				return
			}
		}
	}

	if d.Read != nil && d.Read != d.Create {
		obj, err = r.get(ctx, vals)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read "+d.TypeName+" after create", err.Error())
			return
		}
	}
	if obj, err = r.afterRead(ctx, obj); err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.TypeName, err.Error())
		return
	}
	st, err := r.toState(ctx, obj, plan, modeApply, req.Plan.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.TypeName, err.Error())
		return
	}
	resp.State.Raw = st
}

func (r *frameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := fields(req.State.Raw)
	obj, err := r.get(ctx, state)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err == nil {
		obj, err = r.afterRead(ctx, obj)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+r.d.TypeName, err.Error())
		return
	}
	st, err := r.toState(ctx, obj, state, modeRead, req.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+r.d.TypeName, err.Error())
		return
	}
	resp.State.Raw = st
}

func (r *frameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := fields(req.Plan.Raw)
	state := fields(req.State.Raw)
	d := r.d
	if d.Update == nil {
		resp.Diagnostics.AddError("Update not supported", d.TypeName+" cannot be updated in place; this is a provider bug.")
		return
	}

	include := func(a *Attr) bool { return a.InUpdate }
	if d.Update.Method != http.MethodPut {
		include = func(a *Attr) bool {
			p, s := plan[a.Name], state[a.Name]
			return a.InUpdate && !p.Equal(s)
		}
	}
	// The path comes from the state: ids do not change in place.
	pathVals := Values{}
	for k, v := range state {
		pathVals[k] = v
	}
	for k, v := range plan {
		if v.IsKnown() && !v.IsNull() {
			pathVals[k] = v
		}
	}
	clearRemoved := func(body Object) {
		if d.Update.Method == http.MethodPut {
			return
		}
		// A value removed from configuration is cleared explicitly; leaving
		// it out of a PATCH would keep the old value.
		for _, a := range d.Attrs {
			if a.InUpdate && !a.Local && plan[a.Name].IsNull() && !state[a.Name].IsNull() {
				body[a.JSON] = nil
			}
		}
	}
	var obj Object
	var err error
	if d.Update.Method == http.MethodPut || r.changed(plan, state) {
		obj, err = r.sendBody(ctx, d.Update, plan, "update", include, pathVals, clearRemoved)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update "+d.TypeName, err.Error())
			return
		}
	}
	if d.Read != nil {
		if obj, err = r.get(ctx, pathVals); err != nil {
			resp.Diagnostics.AddError("Unable to read "+d.TypeName+" after update", err.Error())
			return
		}
	}
	if obj, err = r.afterRead(ctx, obj); err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.TypeName, err.Error())
		return
	}
	st, err := r.toState(ctx, obj, plan, modeApply, req.Plan.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.TypeName, err.Error())
		return
	}
	resp.State.Raw = st
}

func (r *frameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	d := r.d
	if d.Delete == nil {
		resp.Diagnostics.AddWarning("Object stays in Telnyx", r.deleteWarning(fields(req.State.Raw)))
		return
	}
	state := fields(req.State.Raw)
	p, err := fillPath(d, d.Delete.Path, state)
	if err == nil {
		_, err = r.c.Send(ctx, d.Delete.Method, p, nil, "", nil)
	}
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete "+d.TypeName, err.Error())
	}
}

func (r *frameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != len(r.d.Import) {
		resp.Diagnostics.AddError("Invalid import id",
			fmt.Sprintf("Expected %s, got %q.", strings.Join(r.d.Import, "/"), req.ID))
		return
	}
	for i, n := range r.d.Import {
		if parts[i] == "" {
			resp.Diagnostics.AddError("Invalid import id", fmt.Sprintf("%s must not be empty.", n))
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(n), types.StringValue(parts[i]))...)
	}
}

func (r *frameworkResource) deleteWarning(state Values) string {
	if r.d.Hooks.DeleteWarning != nil {
		return r.d.Hooks.DeleteWarning(state)
	}
	if r.d.DeleteWarning != "" {
		return r.d.DeleteWarning
	}
	return "The Telnyx API cannot delete this " + r.d.TypeName + ". It was removed from Terraform state but still exists in your Telnyx account."
}

// changed reports whether any attribute the update operation accepts
// differs between plan and state. When only local attributes changed, no
// request is needed.
func (r *frameworkResource) changed(plan, state Values) bool {
	for _, a := range r.d.Attrs {
		if a.InUpdate && !a.Local && !plan[a.Name].Equal(state[a.Name]) {
			return true
		}
	}
	return false
}

// send builds a request body from the selected attributes and sends op.
func (r *frameworkResource) send(ctx context.Context, op *Op, plan Values, name string, include func(*Attr) bool, pathVals Values) (Object, error) {
	return r.sendBody(ctx, op, plan, name, include, pathVals, nil)
}

// sendBody is send with a function that can adjust the body before hooks run.
func (r *frameworkResource) sendBody(ctx context.Context, op *Op, plan Values, name string, include func(*Attr) bool, pathVals Values, adjust func(Object)) (Object, error) {
	body, err := objectToJSON(r.d.Attrs, plan, include)
	if err != nil {
		return nil, err
	}
	if adjust != nil {
		adjust(body)
	}
	if r.d.Hooks.BeforeRequest != nil {
		if err := r.d.Hooks.BeforeRequest(ctx, name, body, plan); err != nil {
			return nil, err
		}
	}
	p, err := fillPath(r.d, op.Path, pathVals)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	out, err := r.c.Send(ctx, op.Method, p, nil, "application/json", data)
	if err != nil {
		return nil, err
	}
	return Decode(out, op.Envelope)
}

// get reads the object with the read operation.
func (r *frameworkResource) get(ctx context.Context, vals Values) (Object, error) {
	return Get(ctx, r.c, r.d, vals)
}

// Get reads an object with the descriptor's read operation. Hooks use it to
// poll.
func Get(ctx context.Context, c *client.Client, d *Resource, vals Values) (Object, error) {
	p, err := fillPath(d, d.Read.Path, vals)
	if err != nil {
		return nil, err
	}
	out, err := c.Send(ctx, d.Read.Method, p, nil, "", nil)
	if err != nil {
		return nil, err
	}
	return Decode(out, d.Read.Envelope)
}

func (r *frameworkResource) afterRead(ctx context.Context, obj Object) (Object, error) {
	if r.d.Hooks.AfterRead == nil {
		return obj, nil
	}
	return r.d.Hooks.AfterRead(ctx, obj)
}

// toState maps an API object to the full state value, carrying over the
// timeouts block from prior.
func (r *frameworkResource) toState(ctx context.Context, obj Object, prior Values, m mode, priorRaw tftypes.Value) (tftypes.Value, error) {
	v, err := objectFromJSON(r.d.Attrs, obj, prior, m)
	if err != nil {
		return tftypes.Value{}, err
	}
	if !r.d.Timeouts {
		return v, nil
	}
	f := fields(v)
	typ := r.d.objectType(ctx)
	if t, ok := fields(priorRaw)[timeoutsAttr]; ok {
		f[timeoutsAttr] = known(t)
	} else {
		f[timeoutsAttr] = tftypes.NewValue(typ.AttributeTypes[timeoutsAttr], nil)
	}
	return tftypes.NewValue(typ, map[string]tftypes.Value(f)), nil
}

// fillPath substitutes {param} placeholders from the attributes that carry
// those path parameters.
func fillPath(d *Resource, p string, vals Values) (string, error) {
	var b strings.Builder
	for {
		i := strings.Index(p, "{")
		if i < 0 {
			b.WriteString(p)
			return b.String(), nil
		}
		j := strings.Index(p[i:], "}")
		if j < 0 {
			return "", fmt.Errorf("malformed path %q", p)
		}
		param := p[i+1 : i+j]
		var attr *Attr
		for _, a := range d.Attrs {
			if slices.Contains(a.PathParams, param) {
				attr = a
			}
		}
		if attr == nil {
			return "", fmt.Errorf("no attribute fills path parameter %s", param)
		}
		v := vals[attr.Name]
		var s string
		if !v.IsKnown() || v.IsNull() || v.As(&s) != nil || s == "" {
			return "", fmt.Errorf("%s is not known", attr.Name)
		}
		b.WriteString(p[:i])
		b.WriteString(url.PathEscape(s))
		p = p[i+j+1:]
	}
}

// fields returns the top-level fields of an object value.
func fields(v tftypes.Value) Values {
	var f map[string]tftypes.Value
	if v.IsNull() || !v.IsKnown() || v.As(&f) != nil {
		return Values{}
	}
	return f
}
