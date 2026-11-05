// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
)

const wave7Spec = `{
  "openapi": "3.1.0",
  "paths": {
    "/providers": {"get": {
      "parameters": [
        {"name": "provider", "in": "query", "schema": {"type": "string"}},
        {"name": "provider_filter", "in": "query", "schema": {"type": "string"}}
      ],
      "responses": {"200": {"content": {"application/json": {"schema": {
        "type": "object",
        "properties": {"data": {"type": "array", "items": {"type": "object", "properties": {"provider": {"type": "string"}}}}}
      }}}}}
    }},
    "/limits": {"get": {
      "parameters": [{"name": "max_items", "in": "query", "schema": {"type": "integer"}}],
      "responses": {"200": {"content": {"application/json": {"schema": {
        "type": "object",
        "properties": {"data": {"type": "array", "items": {"type": "object", "properties": {"id": {"type": "string"}}}}}
      }}}}}
    }},
    "/counts/{count}/items": {"get": {
      "parameters": [{"name": "count", "in": "path", "required": true, "schema": {"type": "string"}}],
      "responses": {"200": {"content": {"application/json": {"schema": {
        "type": "object",
        "properties": {"data": {"type": "array", "items": {"type": "object", "properties": {"id": {"type": "string"}}}}}
      }}}}}
    }},
    "/stt_providers": {"get": {
      "parameters": [{"name": "provider", "in": "query", "schema": {"type": "string"}}],
      "responses": {"200": {"content": {"application/json": {"schema": {
        "type": "object",
        "properties": {"data": {"type": "array", "items": {"type": "object", "properties": {"provider": {"type": "string"}}}}}
      }}}}}
    }}
  }
}`

// A list filter named like a Terraform meta-argument takes the suffix
// _filter and is still sent under its own name.
func TestReservedQueryParam(t *testing.T) {
	dir := t.TempDir()
	sp := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(sp, []byte(wave7Spec), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := apisync.LoadSpec(sp)
	if err != nil {
		t.Fatal(err)
	}
	g := &Generator{Spec: s}
	derive := func(key string) error {
		return deriveParam(t, g, dir, key, "provider")
	}
	if err := derive("GET /stt_providers"); err != nil {
		t.Fatal(err)
	}
	if err := derive("GET /providers"); err == nil {
		t.Error("provider and provider_filter: want a collision error")
	}
	if err := deriveParam(t, g, dir, "GET /limits", "max_items"); err != nil {
		t.Errorf("max_items: %v", err)
	}
	if err := deriveParam(t, g, dir, "GET /counts/{count}/items", ""); err == nil {
		t.Error("reserved path parameter: want an error")
	}
}

// deriveParam derives a list override for key and checks that its only
// parameter is <name>_filter, sent as name.
func deriveParam(t *testing.T, g *Generator, dir, key, name string) error {
	t.Helper()
	p := filepath.Join(dir, "x.yaml")
	_ = os.WriteFile(p, []byte("type: telnyx_xs\nkind: data_source\ndescription: x\nlist: \""+key+"\"\n"), 0o644)
	ov, err := LoadOverride(p)
	if err != nil {
		t.Fatal(err)
	}
	_, l, err := g.DeriveAll(ov)
	if err != nil {
		return err
	}
	if len(l.Params) != 1 || l.Params[0].Name != name+"_filter" || l.Params[0].JSON != name {
		t.Errorf("%s params: %+v", key, l.Params)
	}
	return nil
}
