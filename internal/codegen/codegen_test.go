// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package codegen

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

var update = flag.Bool("update", false, "rewrite golden files")

func fixture(t *testing.T) *Generator {
	t.Helper()
	s, err := apisync.LoadSpec("testdata/spec.json")
	if err != nil {
		t.Fatal(err)
	}
	return &Generator{Spec: s}
}

// TestGolden derives the fixture overrides and compares the output with
// testdata/golden. Run with -update after an intended change.
func TestGolden(t *testing.T) {
	out := t.TempDir()
	rs, err := Run(fixture(t), "testdata/overrides", out, "generated")
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 {
		t.Fatalf("got %d resources", len(rs))
	}
	ctx := context.Background()
	for _, r := range rs {
		if diags := r.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", r.TypeName, diags)
		}
	}
	files, _ := filepath.Glob(filepath.Join(out, "*.go"))
	for _, f := range files {
		got, _ := os.ReadFile(f)
		golden := filepath.Join("testdata", "golden", filepath.Base(f)+".golden")
		if *update {
			if err := os.WriteFile(golden, got, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%v (run go test ./internal/codegen -update)", err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from %s; run go test ./internal/codegen -update and review the diff", filepath.Base(f), golden)
		}
	}
}

func TestRunIsDeterministic(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	for _, dir := range []string{a, b} {
		if _, err := Run(fixture(t), "testdata/overrides", dir, "generated"); err != nil {
			t.Fatal(err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(a, "*.go"))
	for _, f := range files {
		x, _ := os.ReadFile(f)
		y, _ := os.ReadFile(filepath.Join(b, filepath.Base(f)))
		if string(x) != string(y) {
			t.Errorf("%s differs between runs", filepath.Base(f))
		}
	}
}

func TestDerivation(t *testing.T) {
	ov, err := LoadOverride("testdata/overrides/widget.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, err := fixture(t).Derive(ov)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Read.Envelope || !r.Create.Envelope || r.Delete == nil {
		t.Errorf("ops: %+v %+v %+v", r.Create, r.Read, r.Delete)
	}
	checks := map[string]func() bool{
		"name is required":                 func() bool { return r.Attr("name").Required },
		"color enum merges create, update": func() bool { return strings.Join(r.Attr("color").Enum, ",") == "blue,green,red" },
		"region cannot be updated":         func() bool { return r.Attr("region").ForceNew },
		"name can be updated":              func() bool { return !r.Attr("name").ForceNew },
		"nullable size is an int":          func() bool { return r.Attr("size").Kind == kindInt },
		"secret is kept from config":       func() bool { a := r.Attr("secret"); return a.Sensitive && !a.InResponse },
		"labels are JSON":                  func() bool { return r.Attr("labels").Kind == kindJSON },
		"undiscriminated union is JSON":    func() bool { return r.Attr("target").Kind == kindJSON },
		"nullable object is an object":     func() bool { return r.Attr("avatar").Kind == kindObject },
		"arrays of arrays are JSON":        func() bool { return r.Attr("matrix").Kind == kindJSON },
		"status is computed only":          func() bool { a := r.Attr("status"); return a.Computed && !a.Optional },
		"optional returned is O+C":         func() bool { a := r.Attr("color"); return a.Optional && a.Computed },
		"deprecated is marked":             func() bool { return r.Attr("old_flag").Deprecated != "" },
		"renamed":                          func() bool { return r.Attr("widget_count") != nil && r.Attr("widget_count").JSON == "count" },
		"update-only field":                func() bool { a := r.Attr("publish"); return a.InUpdate && !a.InCreate && !a.ForceNew },
		"id fills the path":                func() bool { return r.Attr("id").PathParams[0] == "widget_id" },
		"description is tidied":            func() bool { return r.Attr("name").Description == "The widget's name." },
	}
	for name, ok := range checks {
		if !ok() {
			t.Errorf("%s", name)
		}
	}
	tools := r.Attr("tools")
	if tools.Kind != kindList || len(tools.Attrs) != 3 {
		t.Fatalf("tools: %+v", tools)
	}
	for _, a := range tools.Attrs {
		switch a.Name {
		case "type":
			if strings.Join(a.Enum, ",") != "hangup,webhook" || a.Required {
				t.Errorf("tools.type: %+v", a)
			}
		case "hangup":
			if a.Kind != kindJSON {
				t.Errorf("tools.hangup: %+v", a)
			}
		case "webhook":
			if a.Kind != kindObject {
				t.Errorf("tools.webhook: %+v", a)
			}
		}
	}
	// deep > level2 > level3 are nested attributes; level4, the fourth level
	// of objects, becomes JSON.
	l2 := r.Attr("deep").Attrs[0]
	if l2.Kind != kindObject || l2.Attrs[0].Kind != kindObject || l2.Attrs[0].Attrs[0].Kind != kindJSON {
		t.Errorf("deep nesting: %+v", l2)
	}
}

func TestParentsAndPut(t *testing.T) {
	ov, err := LoadOverride("testdata/overrides/note.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, err := fixture(t).Derive(ov)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(r.Import, "/") != "folder_id/id" {
		t.Errorf("import = %v", r.Import)
	}
	f := r.Attr("folder_id")
	if !f.Required || !f.ForceNew || !f.Local || f.PathParams[0] != "folder_id" {
		t.Errorf("folder_id: %+v", f)
	}
	if c := r.Attr("created_at"); c.Optional || !c.Computed || !c.InUpdate {
		t.Errorf("created_at only in the PUT body must be computed and sent back: %+v", c)
	}
	if s := r.Attr("source_file"); !s.Local || !s.Optional {
		t.Errorf("source_file: %+v", s)
	}
}

func TestOverrideErrors(t *testing.T) {
	g := fixture(t)
	cases := map[string]string{
		"unknown key":       "type: telnyx_x\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\ncolour: red\n",
		"bad type":          "type: widget\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\n",
		"bad operation":     "type: telnyx_x\ndescription: x\ncreate: post /widgets\nread: GET /widgets/{widget_id}\n",
		"missing operation": "type: telnyx_x\ndescription: x\ncreate: POST /gadgets\nread: GET /widgets/{widget_id}\n",
		"unknown attribute": "type: telnyx_x\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\nattributes:\n  nope: {skip: true}\n",
		"reserved name":     "type: telnyx_x\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\nattributes:\n  name: {name: count}\n",
		"duplicate name":    "type: telnyx_x\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\nattributes:\n  color: {name: name}\n",
		"bad exactly_one":   "type: telnyx_x\ndescription: x\ncreate: POST /widgets\nread: GET /widgets/{widget_id}\nexactly_one: [[name, nope]]\n",
	}
	for name, body := range cases {
		p := filepath.Join(t.TempDir(), "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		ov, err := LoadOverride(p)
		if err == nil {
			_, err = g.Derive(ov)
		}
		if err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestWriteRemovesStaleFiles(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "gone.go")
	keep := filepath.Join(dir, "hand_written_test.go")
	_ = os.WriteFile(stale, []byte(Marker+"\npackage generated\n"), 0o644)
	_ = os.WriteFile(keep, []byte("package generated\n"), 0o644)
	if err := Write(dir, map[string][]byte{"new.go": []byte(Marker + "\npackage generated\n")}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("stale generated file must be removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Error("hand-written files must stay")
	}
}

func TestGoName(t *testing.T) {
	for in, want := range map[string]string{
		"telnyx_ai_assistant":                "AIAssistant",
		"telnyx_ai_mcp_server":               "AIMCPServer",
		"telnyx_phone_number_voice_settings": "PhoneNumberVoiceSettings",
	} {
		if got := GoName(in); got != want {
			t.Errorf("GoName(%s) = %s", in, got)
		}
	}
}

const (
	kindInt    = engine.KindInt
	kindJSON   = engine.KindJSON
	kindList   = engine.KindList
	kindObject = engine.KindObject
)
