// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// overrideFrom writes body to a temporary override file and loads it.
func overrideFrom(t *testing.T, body string) *Override {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ov, err := LoadOverride(p)
	if err != nil {
		t.Fatal(err)
	}
	return ov
}

// A field only a create response carries becomes a computed attribute
// marked CreateOnly, found in any 2xx create response.
func TestFromCreate(t *testing.T) {
	g := fixture(t)
	base := "type: telnyx_token\ndescription: x\ncreate: POST /tokens\nread: \"GET /tokens/{id}\"\n"
	r, err := g.Derive(overrideFrom(t, base+"attributes:\n  secret: {sensitive: true, create_only: true}\n"))
	if err != nil {
		t.Fatal(err)
	}
	s := r.Attr("secret")
	if s == nil || !s.CreateOnly || !s.Computed || s.Optional || !s.InResponse || !s.Sensitive {
		t.Errorf("secret: %+v", s)
	}
	if n := r.Attr("name"); n == nil || n.CreateOnly {
		t.Errorf("name: %+v", n)
	}
	if !r.Create.Envelope {
		t.Error("create envelope")
	}

	// Without the key, the field is not an attribute.
	r, err = g.Derive(overrideFrom(t, base))
	if err != nil || r.Attr("secret") != nil {
		t.Errorf("secret without create_only: %v", err)
	}

	for name, attrs := range map[string]string{
		"unknown field": "  nope: {create_only: true}\n",
		"nested":        "  a.b: {create_only: true}\n",
		"in the read":   "  name: {create_only: true}\n",
	} {
		if _, err := g.Derive(overrideFrom(t, base+"attributes:\n"+attrs)); err == nil || !strings.Contains(err.Error(), "create_only") {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// A raw request body needs a Send hook, and a raw read is one value.
func TestRawBodyAndTextRead(t *testing.T) {
	g := fixture(t)
	r, err := g.Derive(overrideFrom(t, "type: telnyx_kv\ndescription: x\n"+
		"create: \"PUT /kv/{key}\"\nread: \"GET /kv/{key}\"\nupdate: \"PUT /kv/{key}\"\ndelete: \"DELETE /kv/{key}\"\n"+
		"id_from_path: key\nattributes:\n  value: {required: true, sensitive: true}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Create.Raw || !r.Update.Raw || !r.Read.Text || r.Read.Envelope {
		t.Errorf("ops: %+v %+v %+v", r.Create, r.Update, r.Read)
	}
	if v := r.Attr("value"); v == nil || !v.Required || !v.InResponse || v.InCreate || v.ForceNew {
		t.Errorf("value: %+v", v)
	}
}

// Union variants of the same shape merge only when the override asks.
func TestMergeVariants(t *testing.T) {
	g := fixture(t)
	base := "type: telnyx_clone\ndescription: x\ncreate: POST /clones\nread: \"GET /clones/{id}\"\ndelete: \"DELETE /clones/{id}\"\n"
	rename := "attributes:\n  provider: {name: clone_provider}\n"
	r, err := g.Derive(overrideFrom(t, base+rename))
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range r.Attrs {
		if a.InCreate {
			t.Errorf("without merge_variants the body is not structured: %+v", a)
		}
	}

	r, err = g.Derive(overrideFrom(t, base+"merge_variants: true\n"+rename))
	if err != nil {
		t.Fatal(err)
	}
	if n := r.Attr("name"); n == nil || !n.Required || !n.InCreate {
		t.Errorf("name, required by every variant: %+v", n)
	}
	if p := r.Attr("clone_provider"); p == nil || !p.Required || strings.Join(p.Enum, ",") != "a,b" {
		t.Errorf("provider, enums joined: %+v", p)
	}
	if m := r.Attr("model"); m == nil || m.Required || m.Enum != nil {
		t.Errorf("model, open in one variant: %+v", m)
	}
	if tr := r.Attr("transcript"); tr == nil || tr.Required || !tr.InCreate || tr.InResponse {
		t.Errorf("transcript, in one variant only: %+v", tr)
	}
}

// Items in results with count are a list paged by page and page_size.
func TestResultsList(t *testing.T) {
	g := fixture(t)
	_, l, err := g.DeriveAll(overrideFrom(t, "type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /result_things\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !l.Op.Results || l.Op.Records || !l.Op.Envelope || l.Op.Paging != engine.PagingPage || l.Op.SizeParam != "page_size" || l.Op.PageSize != 100 {
		t.Errorf("op: %+v", l.Op)
	}
	if len(l.Items.Attrs) != 2 || l.Items.Attrs[0].Name != "id" {
		t.Errorf("items: %+v", l.Items.Attrs)
	}
}

// Boolean and integer query parameters are lookup attributes; other types
// are still errors.
func TestQueryLookupKinds(t *testing.T) {
	g := fixture(t)
	_, err := g.deriveData(overrideFrom(t, "type: telnyx_x\nkind: data_source\ndescription: x\nread: \"GET /lookups/{id}\"\n"))
	if err == nil || !strings.Contains(err.Error(), "query parameter at is a number") {
		t.Fatalf("number query parameter: %v", err)
	}
	g.Spec.Paths()["/lookups/{id}"].(map[string]any)["get"].(map[string]any)["parameters"] = //nolint:forcetypeassert
		g.Spec.Paths()["/lookups/{id}"].(map[string]any)["get"].(map[string]any)["parameters"].([]any)[:3] //nolint:forcetypeassert
	r, err := g.deriveData(overrideFrom(t, "type: telnyx_x\nkind: data_source\ndescription: x\nread: \"GET /lookups/{id}\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if v := r.Attr("verbose"); v == nil || v.Kind != engine.KindBool || v.Query != "verbose" || v.Required {
		t.Errorf("verbose: %+v", v)
	}
	if d := r.Attr("depth"); d == nil || d.Kind != engine.KindInt || d.Query != "depth" {
		t.Errorf("depth: %+v", d)
	}
}

func TestFileNameAvoidsTestFiles(t *testing.T) {
	for in, want := range map[string]string{
		"telnyx_ai_assistant":      "ai_assistant.go",
		"telnyx_ai_assistant_test": "ai_assistant_test_resource.go",
		"telnyx_ai_test_agent":     "ai_test_agent.go",
	} {
		if got := FileName(in); got != want {
			t.Errorf("FileName(%s) = %s", in, got)
		}
	}
}
