// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

const gapsSpec = `{
  "openapi": "3.1.0",
  "paths": {
    "/secrets": {"post": {
      "requestBody": {"content": {"application/json": {"schema": {
        "type": "object", "required": ["name", "value"],
        "properties": {"name": {"type": "string"}, "value": {"type": "string"}}
      }}}},
      "responses": {"201": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"data": {"type": "object", "properties": {
          "name": {"type": "string"}, "value": {"type": "string", "enum": ["REDACTED"]}}}}
      }}}}}
    }},
    "/rooms/{room_id}/token": {"post": {
      "parameters": [{"name": "room_id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"ttl_secs": {"type": "integer"}, "scopes": {"type": "array", "items": {"type": "string"}}}
      }}}},
      "responses": {"201": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"data": {"type": "object", "properties": {"token": {"type": "string"}}}}
      }}}}}
    }},
    "/rooms/{room_id}/grant": {"post": {
      "parameters": [{"name": "room_id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"limits": {"type": "object", "properties": {"max": {"type": "integer"}}}}
      }}}},
      "responses": {"201": {"content": {"application/json": {"schema": {
        "type": "object", "properties": {"data": {"type": "object", "properties": {"token": {"type": "string"}}}}
      }}}}}
    }}
  }
}`

func gapsGenerator(t *testing.T) (*Generator, func(string) *Override) {
	t.Helper()
	dir := t.TempDir()
	sp := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(sp, []byte(gapsSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := apisync.LoadSpec(sp)
	if err != nil {
		t.Fatal(err)
	}
	load := func(body string) *Override {
		p := filepath.Join(dir, "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		ov, err := LoadOverride(p)
		if err != nil {
			t.Fatalf("%v\n%s", err, body)
		}
		return ov
	}
	return &Generator{Spec: s}, load
}

// read: none derives a resource from its create operation alone: no read,
// no data source, attributes from the create response.
func TestReadNone(t *testing.T) {
	g, load := gapsGenerator(t)
	r, err := g.Derive(load(`type: telnyx_secret
description: x
create: POST /secrets
read: none
id: name
attributes:
  name: {force_new: true}
  value: {write_only: true}
`))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Unreadable || r.Read != nil || r.ReadFrom != nil || r.DataSource || r.IDAttr != "name" {
		t.Errorf("descriptor: unreadable %v read %v data source %v id %s", r.Unreadable, r.Read, r.DataSource, r.IDAttr)
	}
	if n := r.Attr("name"); n == nil || !n.Required || !n.ForceNew || !n.InResponse {
		t.Errorf("name: %+v", n)
	}
	if v := r.Attr("value_wo"); v == nil || !v.WriteOnly {
		t.Errorf("value_wo: %+v", v)
	}
	if v := r.Attr("value_wo_version"); v == nil || !v.ForceNew {
		t.Errorf("value_wo_version: %+v", v)
	}

	for name, body := range map[string]string{
		"with update":      "type: telnyx_secret\ndescription: x\ncreate: POST /secrets\nupdate: POST /secrets\nread: none\n",
		"with read_from":   "type: telnyx_secret\ndescription: x\ncreate: POST /secrets\nread: none\nread_from: {op: \"GET /secrets\", key: name}\n",
		"on a data_source": "type: telnyx_secret\nkind: data_source\ndescription: x\nread: none\n",
	} {
		p := filepath.Join(t.TempDir(), "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadOverride(p); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// An ephemeral resource's request fields are optional arguments: not
// computed, not sensitive, sent in the open request.
func TestEphemeralRequestArguments(t *testing.T) {
	g, load := gapsGenerator(t)
	e, err := g.DeriveEphemeral(load("type: telnyx_room_token\nkind: ephemeral\ndescription: x\nopen: \"POST /rooms/{room_id}/token\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*engine.Attr{}
	for _, a := range e.Attrs {
		byName[a.Name] = a
	}
	if ttl := byName["ttl_secs"]; ttl == nil || !ttl.Optional || ttl.Computed || !ttl.InCreate || ttl.InResponse {
		t.Errorf("ttl_secs: %+v", ttl)
	}
	if sc := byName["scopes"]; sc == nil || !sc.Optional || sc.Elem == nil {
		t.Errorf("scopes: %+v", sc)
	}
	if tok := byName["token"]; tok == nil || !tok.Computed || tok.InCreate {
		t.Errorf("token: %+v", tok)
	}
	if _, err := g.DeriveEphemeral(load("type: telnyx_room_grant\nkind: ephemeral\ndescription: x\nopen: \"POST /rooms/{room_id}/grant\"\n")); err == nil {
		t.Error("object request field: want an error")
	}
}
