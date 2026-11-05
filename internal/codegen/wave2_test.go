// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"reflect"
	"testing"
)

// TestConstAndRefinedProperty: a const without a type is a string with one
// allowed value, and an allOf part that narrows a property refines it.
func TestConstAndRefinedProperty(t *testing.T) {
	g := fixture(t)
	s := g.norm(map[string]any{"allOf": []any{
		map[string]any{"type": "object", "properties": map[string]any{
			"kind": map[string]any{"type": "string", "enum": []any{"A", "B"}, "description": "The kind."}}},
		map[string]any{"properties": map[string]any{"kind": map[string]any{"const": "A"}}},
	}})
	kind := g.norm(properties(s)["kind"])
	if kind["type"] != "string" || kind["description"] != "The kind." || !reflect.DeepEqual(stringEnum(kind), []string{"A"}) {
		t.Errorf("refined property: %v", kind)
	}
	if c := g.norm(map[string]any{"const": "EIN"}); c["type"] != "string" {
		t.Errorf("const: %v", c)
	}
}

// TestChildOfFreeFormResponse: a response object without fields echoes the
// request's fields; a map takes its entry schema when that has fields.
func TestChildOfFreeFormResponse(t *testing.T) {
	g := fixture(t)
	entry := map[string]any{"type": "object", "properties": map[string]any{"value": map[string]any{"type": "string"}}}
	req := map[string]any{"type": "object", "properties": map[string]any{"ein": entry}}
	for name, tc := range map[string]struct {
		read map[string]any
		echo bool
	}{
		"free-form":         {map[string]any{"type": "object"}, true},
		"map of union":      {map[string]any{"type": "object", "additionalProperties": map[string]any{"oneOf": []any{entry, map[string]any{"type": "string"}}}}, true},
		"map of objects":    {map[string]any{"type": "object", "additionalProperties": entry}, true},
		"closed":            {map[string]any{"type": "object", "additionalProperties": false}, false},
		"with other fields": {map[string]any{"type": "object", "properties": map[string]any{"other": entry}}, false},
	} {
		c := sources{create: req, read: tc.read}.child("ein", g)
		if (c.read != nil) != tc.echo {
			t.Errorf("%s: read %v", name, c.read)
		}
	}
}

func TestSnake(t *testing.T) {
	for in, want := range map[string]string{
		"webhookURL":          "webhook_url",
		"webhookFailoverURL":  "webhook_failover_url",
		"isTMobileRegistered": "is_t_mobile_registered",
		"brandId":             "brand_id",
		"page[number]":        "page_number",
		"mnoIds":              "mno_ids",
		"sample1":             "sample1",
		"already_snake":       "already_snake",
	} {
		if got := snake(in); got != want {
			t.Errorf("snake(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestAllOfScalar(t *testing.T) {
	g := fixture(t)
	s := g.norm(map[string]any{"allOf": []any{map[string]any{"type": "string", "enum": []any{"A", "B"}}}, "description": "Kind."})
	if s["type"] != "string" || s["description"] != "Kind." || len(stringEnum(s)) != 2 {
		t.Errorf("allOf scalar: %v", s)
	}
}

func TestSameShape(t *testing.T) {
	g := fixture(t)
	str := map[string]any{"type": "string"}
	for name, tc := range map[string]struct {
		a, b map[string]any
		want bool
	}{
		"same":            {str, str, true},
		"int and number":  {map[string]any{"type": "integer"}, map[string]any{"type": "number"}, true},
		"bool and string": {map[string]any{"type": "boolean"}, str, false},
		"untyped":         {map[string]any{}, str, true},
		"strings and objects": {
			map[string]any{"type": "array", "items": str},
			map[string]any{"type": "array", "items": map[string]any{"type": "object"}}, false},
	} {
		if got := sameShape(g, tc.a, tc.b); got != tc.want {
			t.Errorf("%s: %v", name, got)
		}
	}
}
