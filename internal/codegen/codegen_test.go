// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

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
	gen, err := Run(fixture(t), "testdata/overrides", out, "generated")
	if err != nil {
		t.Fatal(err)
	}
	rs, ls := gen.Resources, gen.Lists
	if len(rs) != 7 || len(ls) != 2 || len(gen.Actions) != 1 || len(gen.Ephemerals) != 1 {
		t.Fatalf("got %d resources, %d lists, %d actions and %d ephemeral resources", len(rs), len(ls), len(gen.Actions), len(gen.Ephemerals))
	}
	ctx := context.Background()
	for _, r := range rs {
		if !r.NoResource {
			if diags := r.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
				t.Errorf("%s: %v", r.TypeName, diags)
			}
		}
		if r.DataSource {
			if diags := r.DataSourceSchema().ValidateImplementation(ctx); diags.HasError() {
				t.Errorf("%s data source: %v", r.TypeName, diags)
			}
		}
	}
	for _, l := range ls {
		if diags := l.Schema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", l.TypeName, diags)
		}
	}
	for _, a := range gen.Actions {
		if diags := a.Schema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", a.TypeName, diags)
		}
	}
	for _, e := range gen.Ephemerals {
		if diags := e.Schema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", e.TypeName, diags)
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

func TestCollectionShapes(t *testing.T) {
	g := fixture(t)
	load := func(name string) *Override {
		ov, err := LoadOverride("testdata/overrides/" + name + ".yaml")
		if err != nil {
			t.Fatal(err)
		}
		return ov
	}
	r, l, err := g.DeriveAll(load("gadget"))
	if err != nil {
		t.Fatal(err)
	}
	if l.TypeName != "telnyx_gadgets" || l.Op.Paging != engine.PagingNumber || l.Op.PageSize != 100 {
		t.Errorf("list: %+v %+v", l, l.Op)
	}
	var params []string
	for _, p := range l.Params {
		params = append(params, p.Name+"="+p.JSON)
	}
	if got := strings.Join(params, " "); got != "filter_name=filter[name][eq] filter_name_contains=filter[name][contains] filter_status=filter[status] filter_tag=filter[tag] sort=sort" {
		t.Errorf("params: %s", got)
	}
	if u := r.Updates; len(u) != 1 || strings.Join(u[0].Attrs, ",") != "blocklist_id" || u[0].Clear == nil {
		t.Errorf("updates: %+v", u)
	}
	if b := r.Attr("blocklist_id"); !b.Optional || b.Computed || b.ForceNew || b.InUpdate || b.InCreate {
		t.Errorf("blocklist_id: %+v", b)
	}

	tag, _, err := g.DeriveAll(load("gadget_tag"))
	if err != nil {
		t.Fatal(err)
	}
	if a := tag.Attr("tag"); tag.IDAttr != "tag" || !a.Required || !a.ForceNew || !a.InCreate || tag.DataSource {
		t.Errorf("tag: %+v %+v", tag, a)
	}
	if strings.Join(tag.Import, "/") != "gadget_id/tag" || tag.ReadFrom.Field != "tags" || tag.Read != nil {
		t.Errorf("tag: %+v", tag)
	}

	sec, _, err := g.DeriveAll(load("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if sec.ReadFrom.Op.Paging != engine.PagingFlat || sec.ReadFrom.Op.PageSize != 50 || sec.ReadFrom.Key != "id" {
		t.Errorf("secret read_from: %+v %+v", sec.ReadFrom, sec.ReadFrom.Op)
	}

	bal, l, err := g.DeriveAll(load("balance"))
	if err != nil || l != nil || !bal.NoResource || !bal.DataSource || !bal.Attr("currency").Computed {
		t.Errorf("balance: %+v %+v %v", bal, l, err)
	}
	reg, l, err := g.DeriveAll(load("regions"))
	if err != nil || reg != nil || l.TypeName != "telnyx_regions" || l.Items.Elem.Kind != engine.KindString ||
		l.Op.Paging != engine.PagingNumber || l.Op.PageSize != 10 {
		t.Errorf("regions: %+v %+v %v", reg, l, err)
	}
}

func TestCollectionErrors(t *testing.T) {
	g := fixture(t)
	cases := map[string]string{
		"unsupported paging":        "type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /offset_things\n",
		"read and read_from":        "type: telnyx_x\ndescription: x\ncreate: POST /secrets\nread: GET /widgets/{widget_id}\nread_from: {op: GET /secrets, key: id}\n",
		"object elements, no key":   "type: telnyx_x\ndescription: x\ncreate: POST /secrets\nread_from: {op: GET /secrets}\n",
		"plain values with key":     "type: telnyx_x\ndescription: x\ncreate: POST /gadgets/{gadget_id}/tags\nid: tag\nread_from: {op: \"GET /gadgets/{gadget_id}\", field: tags, key: id}\n",
		"plain values, no id":       "type: telnyx_x\ndescription: x\ncreate: POST /gadgets/{gadget_id}/tags\nread_from: {op: \"GET /gadgets/{gadget_id}\", field: tags}\n",
		"not an array field":        "type: telnyx_x\ndescription: x\ncreate: POST /gadgets/{gadget_id}/tags\nid: tag\nread_from: {op: \"GET /gadgets/{gadget_id}\", field: name}\n",
		"update without body":       "type: telnyx_x\ndescription: x\ncreate: POST /gadgets\nread: GET /gadgets/{gadget_id}\nupdates: [{op: \"POST /gadgets/{gadget_id}/actions/remove_blocklist\"}]\n",
		"data source with create":   "type: telnyx_x\nkind: data_source\ndescription: x\ncreate: POST /gadgets\nread: GET /gadgets/{gadget_id}\n",
		"bad paging value":          "type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /regions\npaging: cursor\n",
		"list_name without list":    "type: telnyx_x\ndescription: x\ncreate: POST /gadgets\nread: GET /gadgets/{gadget_id}\nlist_name: telnyx_xs\n",
		"singleton, id in response": "type: telnyx_x\ndescription: x\ncreate: POST /gadgets\nread: GET /gadgets/{gadget_id}\nsingleton: true\n",
		"singleton with id":         "type: telnyx_x\ndescription: x\ncreate: PATCH /settings\nread: GET /settings\nsingleton: true\nid: name\n",
		"list not returning array":  "type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /balance\n",
	}
	for name, body := range cases {
		p := filepath.Join(t.TempDir(), "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		ov, err := LoadOverride(p)
		if err == nil {
			_, _, err = g.DeriveAll(ov)
		}
		if err == nil {
			t.Errorf("%s: want an error", name)
		} else {
			t.Logf("%s: %v", name, err)
		}
	}

	// paging: none reads one page of a list the engine cannot page.
	p := filepath.Join(t.TempDir(), "x.yaml")
	_ = os.WriteFile(p, []byte("type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /offset_things\npaging: none\n"), 0o644)
	ov, err := LoadOverride(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, l, err := g.DeriveAll(ov); err != nil || l.Op.Paging != engine.PagingNone {
		t.Errorf("paging none: %+v %v", l, err)
	}

	// A singleton has a fixed id and no path parameters.
	_ = os.WriteFile(p, []byte("type: telnyx_settings\ndescription: x\ncreate: PATCH /settings\nread: GET /settings\nupdate: PATCH /settings\nsingleton: true\n"), 0o644)
	if ov, err = LoadOverride(p); err != nil {
		t.Fatal(err)
	}
	if r, _, err := g.DeriveAll(ov); err != nil || r.FixedID != "settings" || r.IDAttr != "id" || !r.Attr("id").Computed || !r.Attr("channels").InCreate {
		t.Errorf("singleton: %+v %v", r, err)
	}

	// page[after] without a page number pages by cursor.
	_ = os.WriteFile(p, []byte("type: telnyx_x\nkind: data_source\ndescription: x\nlist: GET /cursor_things\n"), 0o644)
	if ov, err = LoadOverride(p); err != nil {
		t.Fatal(err)
	}
	if _, l, err := g.DeriveAll(ov); err != nil || l.Op.Paging != engine.PagingCursor {
		t.Errorf("cursor paging: %+v %v", l, err)
	}
}

func TestPlural(t *testing.T) {
	for in, want := range map[string]string{
		"telnyx_billing_group":    "telnyx_billing_groups",
		"telnyx_address":          "telnyx_addresses",
		"telnyx_policy":           "telnyx_policies",
		"telnyx_gateway":          "telnyx_gateways",
		"telnyx_voice_settings":   "telnyx_voice_settings",
		"telnyx_messaging_box":    "telnyx_messaging_boxes",
		"telnyx_ai_assistant_tag": "telnyx_ai_assistant_tags",
	} {
		if got := Plural(in); got != want {
			t.Errorf("Plural(%s) = %s", in, got)
		}
	}
}

func TestPart2Shapes(t *testing.T) {
	g := fixture(t)
	load := func(name string) *Override {
		ov, err := LoadOverride("testdata/overrides/" + name + ".yaml")
		if err != nil {
			t.Fatal(err)
		}
		return ov
	}
	r, err := g.Derive(load("vault"))
	if err != nil {
		t.Fatal(err)
	}
	if pw := r.Attr("password_wo"); pw == nil || !pw.WriteOnly || !pw.Sensitive || pw.Computed || pw.InResponse || pw.Version != "password_wo_version" {
		t.Errorf("password_wo: %+v", pw)
	}
	if r.Attr("password") != nil {
		t.Error("the stored password attribute must not exist")
	}
	if v := r.Attr("password_wo_version"); v == nil || !v.Local || v.ForceNew || v.Kind != engine.KindInt {
		t.Errorf("password_wo_version: %+v", v)
	}
	if rel := r.Attr("release_on_destroy"); !r.ReleaseOnDestroy || rel == nil || !rel.Local || rel.ForceNew {
		t.Errorf("release_on_destroy: %+v", rel)
	}

	a, err := g.DeriveAction(load("gadget_set_blocklist"))
	if err != nil {
		t.Fatal(err)
	}
	if id := a.Attrs[1]; id.Name != "gadget_id" || !id.Required || !id.Local {
		t.Errorf("action path parameter: %+v", id)
	}
	if b := a.Attrs[0]; b.Name != "blocklist_id" || b.Required || !b.Optional || b.Computed {
		t.Errorf("action body field: %+v", b)
	}

	e, err := g.DeriveEphemeral(load("vault_token"))
	if err != nil {
		t.Fatal(err)
	}
	if e.Open.Method != "POST" || !e.Open.Envelope || len(e.Attrs) != 3 || !e.Attrs[1].Computed {
		t.Errorf("ephemeral: %+v %+v", e.Open, e.Attrs)
	}
}

func TestPart2Errors(t *testing.T) {
	g := fixture(t)
	vault := "type: telnyx_x\ndescription: x\ncreate: POST /vaults\nread: \"GET /vaults/{vault_id}\"\n"
	cases := map[string]string{
		"release without delete":     vault + "release_on_destroy: true\n",
		"write_only nested":          vault + "attributes:\n  settings.pin: {write_only: true}\n",
		"write_only on a read field": vault + "attributes:\n  id: {write_only: true}\n",
		"action without op":          "type: telnyx_x\nkind: action\ndescription: x\n",
		"action with create":         "type: telnyx_x\nkind: action\ndescription: x\nop: POST /vaults\ncreate: POST /vaults\n",
		"op on a resource":           vault + "op: POST /vaults\n",
		"ephemeral with op":          "type: telnyx_x\nkind: ephemeral\ndescription: x\nop: POST /vaults\n",
		"ephemeral without JSON":     "type: telnyx_x\nkind: ephemeral\ndescription: x\nopen: \"DELETE /vaults/{vault_id}\"\n",
		"action path clash":          "type: telnyx_x\nkind: action\ndescription: x\nop: \"POST /gadgets/{gadget_id}/actions/set_blocklist\"\npath_params: {gadget_id: blocklist_id}\n",
		"unknown kind":               "type: telnyx_x\nkind: thing\ndescription: x\n",
	}
	for name, body := range cases {
		p := filepath.Join(t.TempDir(), "x.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		ov, err := LoadOverride(p)
		if err == nil {
			switch ov.Kind {
			case "action":
				_, err = g.DeriveAction(ov)
			case "ephemeral":
				_, err = g.DeriveEphemeral(ov)
			default:
				_, err = g.Derive(ov)
			}
		}
		if err == nil {
			t.Errorf("%s: want an error", name)
		} else {
			t.Logf("%s: %v", name, err)
		}
	}
}
