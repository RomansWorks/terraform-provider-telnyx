// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// DeriveAction derives an action from an override with kind: action. Its
// attributes are the path parameters and the request body fields.
func (g *Generator) DeriveAction(ov *Override) (*engine.Action, error) {
	op, err := g.operation(ov.Op)
	if err != nil {
		return nil, fmt.Errorf("%s: op: %w", ov.Type, err)
	}
	d := &deriver{g: g, ov: ov, used: map[string]bool{}}
	var attrs []*engine.Attr
	if body := g.requestSchema(op); body != nil {
		if attrs, err = d.attrs(sources{create: body}, 0, false, ""); err != nil {
			return nil, fmt.Errorf("%s: %w", ov.Type, err)
		}
	}
	for _, a := range attrs {
		if a.WriteOnly {
			return nil, fmt.Errorf("%s: an action stores nothing, so write_only does not apply", ov.Type)
		}
		a.Computed, a.ForceNew, a.InResponse = false, false, false
		a.Optional = !a.Required
	}
	a := &engine.Action{TypeName: ov.Type, Description: ov.Description,
		Op: &engine.Op{Method: op.method, Path: op.path}, Attrs: attrs}
	if err := addPathParams(ov, op, &a.Attrs); err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	if err := checkUnused(ov, d); err != nil {
		return nil, err
	}
	sort.SliceStable(a.Attrs, func(i, j int) bool { return a.Attrs[i].Name < a.Attrs[j].Name })
	return a, nil
}

// DeriveEphemeral derives an ephemeral resource from an override with
// kind: ephemeral. Path parameters are required; response fields computed.
func (g *Generator) DeriveEphemeral(ov *Override) (*engine.Ephemeral, error) {
	op, err := g.operation(ov.Open)
	if err != nil {
		return nil, fmt.Errorf("%s: open: %w", ov.Type, err)
	}
	schema, env := g.responseSchema(op)
	text := false
	if schema == nil && g.textResponse(op) {
		// A plain-text answer, such as a token, is one string field.
		schema = map[string]any{"type": "object", "properties": map[string]any{
			"value": map[string]any{"type": "string", "readOnly": true, "description": "The value the API returned."}}}
		text = true
	}
	if schema == nil {
		return nil, fmt.Errorf("%s: %s has no JSON or plain-text response schema", ov.Type, ov.Open)
	}
	d := &deriver{g: g, ov: ov, used: map[string]bool{}}
	attrs, err := d.attrs(sources{read: schema}, 0, false, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	e := &engine.Ephemeral{TypeName: ov.Type, Description: ov.Description,
		Open: &engine.Op{Method: op.method, Path: op.path, Envelope: env, Text: text}, Attrs: attrs}
	if err := addPathParams(ov, op, &e.Attrs); err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	if err := checkUnused(ov, d); err != nil {
		return nil, err
	}
	sort.SliceStable(e.Attrs, func(i, j int) bool { return e.Attrs[i].Name < e.Attrs[j].Name })
	return e, nil
}

// addPathParams adds one required local attribute per path parameter.
func addPathParams(ov *Override, op *operation, attrs *[]*engine.Attr) error {
	for _, p := range pathParams(op.path) {
		name := snake(p)
		if n, ok := ov.PathParams[p]; ok {
			name = n
		}
		for _, a := range *attrs {
			if a.Name == name {
				return fmt.Errorf("path parameter %s and field %s both map to %s; set path_params", p, a.JSON, name)
			}
		}
		*attrs = append(*attrs, &engine.Attr{Name: name, Kind: engine.KindString, Required: true, Local: true,
			PathParams:  []string{p},
			Description: paramDescription(name)})
	}
	return nil
}

// paramDescription describes a path parameter attribute.
func paramDescription(name string) string {
	if name == "id" {
		return "The object's id."
	}
	return "The " + strings.ReplaceAll(strings.TrimSuffix(name, "_id"), "_", " ") + " id."
}
