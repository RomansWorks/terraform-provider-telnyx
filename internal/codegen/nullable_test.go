// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
)

const nullableSpec = `{
  "openapi": "3.1.0",
  "paths": {
    "/numbers/{id}/actions/bundle": {"patch": {
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {
        "type": "object", "required": ["bundle_id"],
        "properties": {"bundle_id": {"type": ["string", "null"]}}
      }}}},
      "responses": {"200": {"description": "ok"}}
    }},
    "/numbers/{id}": {"get": {
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "responses": {"200": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"data": {"type": "object", "properties": {"bundle_id": {"type": "string"}}}}
      }}}}}
    }}
  }
}`

// A nullable action field is optional and marked to be sent as null when
// unset; on anything but an action request it is an error.
func TestNullable(t *testing.T) {
	dir := t.TempDir()
	sp := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(sp, []byte(nullableSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := apisync.LoadSpec(sp)
	if err != nil {
		t.Fatal(err)
	}
	g := &Generator{Spec: s}
	load := func(body string) *Override {
		p := filepath.Join(dir, "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		ov, err := LoadOverride(p)
		if err != nil {
			t.Fatal(err)
		}
		return ov
	}
	a, err := g.DeriveAction(load(`type: telnyx_number_bundle
kind: action
description: x
op: "PATCH /numbers/{id}/actions/bundle"
attributes:
  bundle_id: {nullable: true}
`))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, at := range a.Attrs {
		if at.Name == "bundle_id" {
			found = true
			if !at.Nullable || at.Required || !at.Optional {
				t.Errorf("bundle_id: %+v", at)
			}
		}
	}
	if !found {
		t.Fatal("no bundle_id attribute")
	}
	if _, _, err := g.DeriveAll(load(`type: telnyx_number
kind: data_source
description: x
read: "GET /numbers/{id}"
attributes:
  bundle_id: {nullable: true}
`)); err == nil {
		t.Error("nullable on a data source: want an error")
	}
}
