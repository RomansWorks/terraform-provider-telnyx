// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type (
	planModifierString  = planmodifier.String
	planModifierInt64   = planmodifier.Int64
	planModifierFloat64 = planmodifier.Float64
	planModifierBool    = planmodifier.Bool
	planModifierList    = planmodifier.List
	planModifierSet     = planmodifier.Set
	planModifierObject  = planmodifier.Object
	attrTypeT           = attr.Type
)

// timeoutsAttr is the name of the optional timeouts block.
const timeoutsAttr = "timeouts"

// ResourceSchema builds the resource schema from the descriptor.
func (r *Resource) ResourceSchema(ctx context.Context) schema.Schema {
	attrs := make(map[string]schema.Attribute, len(r.Attrs)+1)
	for _, a := range r.Attrs {
		attrs[a.Name] = resourceAttr(a)
	}
	if r.Timeouts {
		attrs[timeoutsAttr] = timeouts.Attributes(ctx, timeouts.Opts{Create: true})
	}
	return schema.Schema{MarkdownDescription: r.Description, Attributes: attrs}
}

// objectType returns the full tftypes type of the resource's state,
// including the timeouts block when present.
func (r *Resource) objectType(ctx context.Context) tftypes.Object {
	t := ObjectType(r.Attrs)
	if r.Timeouts {
		t.AttributeTypes[timeoutsAttr] = r.ResourceSchema(ctx).Attributes[timeoutsAttr].GetType().TerraformType(ctx)
	}
	return t
}

func description(a *Attr) string {
	d := a.Description
	if len(a.Enum) > 0 {
		quoted := make([]string, len(a.Enum))
		for i, e := range a.Enum {
			quoted[i] = "`" + e + "`"
		}
		d = strings.TrimSpace(d + " One of " + strings.Join(quoted, ", ") + ".")
	}
	if a.Kind == KindJSON {
		d = strings.TrimSpace(d + " A JSON document; use `jsonencode()`.")
	}
	if !a.InResponse && !a.Local && !a.Computed {
		d = strings.TrimSpace(d + " The API does not return this value, so changes made outside Terraform are not detected.")
	}
	return d
}

func resourceAttr(a *Attr) schema.Attribute {
	req, opt, comp := a.Required, a.Optional, a.Computed
	desc := description(a)
	var stringValidators []validator.String
	if len(a.Enum) > 0 && (a.Kind == KindString) {
		stringValidators = append(stringValidators, stringvalidator.OneOf(a.Enum...))
	}
	keep := comp // keep computed values stable across plans
	switch a.Kind {
	case KindString, KindJSON:
		var pm []planModifierString
		if keep {
			pm = append(pm, stringplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, stringplanmodifier.RequiresReplace())
		}
		return schema.StringAttribute{Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive,
			MarkdownDescription: desc, Validators: stringValidators, PlanModifiers: pm}
	case KindInt:
		var pm []planModifierInt64
		if keep {
			pm = append(pm, int64planmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, int64planmodifier.RequiresReplace())
		}
		return schema.Int64Attribute{Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive,
			MarkdownDescription: desc, PlanModifiers: pm}
	case KindFloat:
		var pm []planModifierFloat64
		if keep {
			pm = append(pm, float64planmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, float64planmodifier.RequiresReplace())
		}
		return schema.Float64Attribute{Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive,
			MarkdownDescription: desc, PlanModifiers: pm}
	case KindBool:
		var pm []planModifierBool
		if keep {
			pm = append(pm, boolplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, boolplanmodifier.RequiresReplace())
		}
		return schema.BoolAttribute{Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive,
			MarkdownDescription: desc, PlanModifiers: pm}
	case KindList:
		var pm []planModifierList
		if keep {
			pm = append(pm, listplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, listplanmodifier.RequiresReplace())
		}
		if a.Elem != nil {
			return schema.ListAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt, Computed: comp,
				Sensitive: a.Sensitive, MarkdownDescription: desc, PlanModifiers: pm}
		}
		return schema.ListNestedAttribute{NestedObject: schema.NestedAttributeObject{Attributes: nested(a.Attrs)},
			Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive, MarkdownDescription: desc, PlanModifiers: pm}
	case KindSet:
		var pm []planModifierSet
		if keep {
			pm = append(pm, setplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, setplanmodifier.RequiresReplace())
		}
		if a.Elem != nil {
			return schema.SetAttribute{ElementType: attrType(a.Elem), Required: req, Optional: opt, Computed: comp,
				Sensitive: a.Sensitive, MarkdownDescription: desc, PlanModifiers: pm}
		}
		return schema.SetNestedAttribute{NestedObject: schema.NestedAttributeObject{Attributes: nested(a.Attrs)},
			Required: req, Optional: opt, Computed: comp, Sensitive: a.Sensitive, MarkdownDescription: desc, PlanModifiers: pm}
	case KindObject:
		var pm []planModifierObject
		if keep {
			pm = append(pm, objectplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			pm = append(pm, objectplanmodifier.RequiresReplace())
		}
		return schema.SingleNestedAttribute{Attributes: nested(a.Attrs), Required: req, Optional: opt, Computed: comp,
			Sensitive: a.Sensitive, MarkdownDescription: desc, PlanModifiers: pm}
	}
	panic(fmt.Sprintf("engine: unknown kind for %s", a.Name))
}

func nested(attrs []*Attr) map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(attrs))
	for _, a := range attrs {
		out[a.Name] = resourceAttr(a)
	}
	return out
}

// attrType returns the framework type of a scalar list or set element.
func attrType(a *Attr) attrTypeT {
	switch a.Kind {
	case KindInt:
		return types.Int64Type
	case KindFloat:
		return types.Float64Type
	case KindBool:
		return types.BoolType
	}
	return types.StringType
}

// DataSourceSchema builds the data source schema: lookup attributes are
// required, everything else computed.
func (r *Resource) DataSourceSchema() dschema.Schema {
	lookup := map[string]bool{}
	for _, n := range r.Import {
		lookup[n] = true
	}
	attrs := map[string]dschema.Attribute{}
	for _, a := range r.Attrs {
		if a.Local || (!a.InResponse && !lookup[a.Name]) {
			continue
		}
		attrs[a.Name] = dataAttr(a, lookup[a.Name])
	}
	return dschema.Schema{MarkdownDescription: r.Description, Attributes: attrs}
}

// dataAttrs returns the descriptor attributes present in the data source.
func (r *Resource) dataAttrs() []*Attr {
	lookup := map[string]bool{}
	for _, n := range r.Import {
		lookup[n] = true
	}
	var out []*Attr
	for _, a := range r.Attrs {
		if a.Local || (!a.InResponse && !lookup[a.Name]) {
			continue
		}
		c := *a
		c.InResponse = a.InResponse || !lookup[a.Name]
		out = append(out, &c)
	}
	return out
}

func dataAttr(a *Attr, required bool) dschema.Attribute {
	desc := description(a)
	comp := !required
	switch a.Kind {
	case KindString, KindJSON:
		return dschema.StringAttribute{Required: required, Computed: comp, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindInt:
		return dschema.Int64Attribute{Computed: comp, Required: required, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindFloat:
		return dschema.Float64Attribute{Computed: comp, Required: required, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindBool:
		return dschema.BoolAttribute{Computed: comp, Required: required, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindList:
		if a.Elem != nil {
			return dschema.ListAttribute{ElementType: attrType(a.Elem), Computed: true, Sensitive: a.Sensitive, MarkdownDescription: desc}
		}
		return dschema.ListNestedAttribute{NestedObject: dschema.NestedAttributeObject{Attributes: dataNested(a.Attrs)},
			Computed: true, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindSet:
		if a.Elem != nil {
			return dschema.SetAttribute{ElementType: attrType(a.Elem), Computed: true, Sensitive: a.Sensitive, MarkdownDescription: desc}
		}
		return dschema.SetNestedAttribute{NestedObject: dschema.NestedAttributeObject{Attributes: dataNested(a.Attrs)},
			Computed: true, Sensitive: a.Sensitive, MarkdownDescription: desc}
	case KindObject:
		return dschema.SingleNestedAttribute{Attributes: dataNested(a.Attrs), Computed: true, Sensitive: a.Sensitive, MarkdownDescription: desc}
	}
	panic(fmt.Sprintf("engine: unknown kind for %s", a.Name))
}

func dataNested(attrs []*Attr) map[string]dschema.Attribute {
	out := make(map[string]dschema.Attribute, len(attrs))
	for _, a := range attrs {
		out[a.Name] = dataAttr(a, false)
	}
	return out
}
