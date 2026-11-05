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
		if len(l.Params) != 1 || l.Params[0].Name != "provider_filter" || l.Params[0].JSON != "provider" {
			t.Errorf("params: %+v", l.Params)
		}
		return nil
	}
	if err := derive("GET /stt_providers"); err != nil {
		t.Fatal(err)
	}
	if err := derive("GET /providers"); err == nil {
		t.Error("provider and provider_filter: want a collision error")
	}
}
