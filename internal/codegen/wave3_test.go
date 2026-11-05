// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
)

// wave3Spec has a tunnel whose token only create returns, a peer whose
// private key read also lists, and a card whose status object is
// read-only field by field.
const wave3Spec = `{"openapi": "3.1.0", "paths": {
 "/tunnels": {"post": {
  "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}}}},
  "responses": {"201": {"content": {"application/json": {"schema": {"type": "object", "properties": {"data": {"allOf": [
   {"$ref": "#/components/schemas/Tunnel"},
   {"type": "object", "properties": {"tunnel_token": {"type": "string", "readOnly": true, "description": "The token."}}}]}}}}}}}}},
 "/tunnels/{id}": {
  "get": {"parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
   "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"data": {"$ref": "#/components/schemas/Tunnel"}}}}}}}},
  "patch": {"parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
   "requestBody": {"content": {"application/json": {"schema": {"type": "object", "properties": {
    "name": {"type": "string"},
    "status": {"type": "object", "properties": {"value": {"type": "string", "readOnly": true}, "reason": {"type": "string", "readOnly": true}}},
    "limit": {"type": "object", "properties": {"amount": {"type": "string"}, "unit": {"type": "string", "readOnly": true}}}}}}}},
   "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"data": {"$ref": "#/components/schemas/Tunnel"}}}}}}}}}},
 "components": {"schemas": {"Tunnel": {"type": "object", "properties": {
  "id": {"type": "string", "readOnly": true},
  "name": {"type": "string"},
  "private_key": {"type": "string", "readOnly": true},
  "status": {"type": "object", "properties": {"value": {"type": "string"}, "reason": {"type": "string"}}},
  "limit": {"type": "object", "properties": {"amount": {"type": "string"}, "unit": {"type": "string"}}}}}}}}`

func wave3(t *testing.T, override string) (*Generator, *Override) {
	t.Helper()
	dir := t.TempDir()
	sp := filepath.Join(dir, "spec.json")
	op := filepath.Join(dir, "x.yaml")
	if err := os.WriteFile(sp, []byte(wave3Spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(op, []byte(override), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := apisync.LoadSpec(sp)
	if err != nil {
		t.Fatal(err)
	}
	ov, err := LoadOverride(op)
	if err != nil {
		t.Fatal(err)
	}
	return &Generator{Spec: s}, ov
}

const tunnelOverride = "type: telnyx_tunnel\ndescription: x\ncreate: POST /tunnels\nread: \"GET /tunnels/{id}\"\nupdate: \"PATCH /tunnels/{id}\"\n"

func TestCreateOnly(t *testing.T) {
	g, ov := wave3(t, tunnelOverride+"attributes:\n  tunnel_token: {create_only: true}\n  private_key: {create_only: true}\n")
	r, err := g.Derive(ov)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"tunnel_token", "private_key"} {
		a := r.Attr(n)
		if a == nil || !a.CreateOnly || !a.Computed || a.Optional || !a.Sensitive || !a.InResponse || a.InCreate || a.InUpdate {
			t.Errorf("%s: %+v", n, a)
		}
	}
	for name, attrs := range map[string]string{
		"configurable": "  name: {create_only: true}\n",
		"nowhere":      "  missing: {create_only: true}\n",
		"nested":       "  status.value: {create_only: true}\n",
	} {
		g, ov := wave3(t, tunnelOverride+"attributes:\n"+attrs)
		if _, err := g.Derive(ov); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// TestReadOnlyObject: an object whose request fields are all read-only is
// not configurable; one with any writable field is.
func TestReadOnlyObject(t *testing.T) {
	g, ov := wave3(t, tunnelOverride)
	r, err := g.Derive(ov)
	if err != nil {
		t.Fatal(err)
	}
	if s := r.Attr("status"); s == nil || s.InUpdate || s.Optional || !s.Computed {
		t.Errorf("status: %+v", s)
	}
	if l := r.Attr("limit"); l == nil || !l.InUpdate || !l.Optional {
		t.Errorf("limit: %+v", l)
	}
}
