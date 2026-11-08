// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// deriveYAML derives one override written as YAML.
func deriveYAML(t *testing.T, body string) (*engine.Resource, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ov, err := LoadOverride(p)
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := fixture(t).DeriveAll(ov)
	return r, err
}

// A limit is identified by its product and period: read from the list by
// both, period sent as a query parameter on update and delete, and the
// create body's variants told apart by a boolean merged into one object.
func TestMatchQueryAndFlagUnion(t *testing.T) {
	r, err := deriveYAML(t, `type: telnyx_limit
description: x
create: POST /limits
update: PATCH /limits/{product}
delete: DELETE /limits/{product}
id: product
read_from: {op: "GET /limits", key: product, match: [period]}
query: [period]
`)
	if err != nil {
		t.Fatal(err)
	}
	if r.IDAttr != "product" || !slices.Equal(r.Import, []string{"product", "period"}) {
		t.Errorf("id %s, import %v", r.IDAttr, r.Import)
	}
	if !slices.Equal(r.ReadFrom.Match, []string{"period"}) {
		t.Errorf("match %v", r.ReadFrom.Match)
	}
	if p := r.Attr("period"); p == nil || !p.Required || !p.ForceNew || p.InUpdate {
		t.Errorf("period %+v", p)
	}
	if !slices.Equal(r.Update.Query, []string{"period"}) || !slices.Equal(r.Delete.Query, []string{"period"}) {
		t.Errorf("query: update %v, delete %v", r.Update.Query, r.Delete.Query)
	}
	// One variant requires amount, the other unlimited: neither is
	// required, and unlimited is a plain boolean.
	if a := r.Attr("amount"); a == nil || a.Required || !a.InCreate || a.Kind != engine.KindFloat {
		t.Errorf("amount %+v", a)
	}
	if u := r.Attr("unlimited"); u == nil || u.Required || u.Kind != engine.KindBool || len(u.Enum) > 0 {
		t.Errorf("unlimited %+v", u)
	}

	for name, body := range map[string]string{
		"match unknown": "type: telnyx_limit\ndescription: x\ncreate: POST /limits\nid: product\nread_from: {op: \"GET /limits\", key: product, match: [nope]}\n",
		"query unknown": "type: telnyx_limit\ndescription: x\ncreate: POST /limits\nupdate: PATCH /limits/{product}\nid: product\nread_from: {op: \"GET /limits\", key: product}\nquery: [nope]\n",
		"query unused":  "type: telnyx_limit\ndescription: x\ncreate: POST /limits\nupdate: PATCH /limits/{product}\nid: product\nread_from: {op: \"GET /limits\", key: product}\nquery: [spend]\n",
	} {
		if _, err := deriveYAML(t, body); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// A once-per-account object whose response has its own id gets the fixed
// id in its place.
func TestSingletonReplacesResponseID(t *testing.T) {
	r, err := deriveYAML(t, "type: telnyx_profile\ndescription: x\ncreate: PATCH /profile\nread: GET /profile\nupdate: PATCH /profile\nsingleton: true\n")
	if err != nil {
		t.Fatal(err)
	}
	id := r.Attr("id")
	if r.FixedID != "profile" || id == nil || id.Kind != engine.KindString || !id.Computed {
		t.Errorf("fixed id %q, id %+v", r.FixedID, id)
	}
	n := 0
	for _, a := range r.Attrs {
		if a.Name == "id" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d id attributes", n)
	}
}

// Members are added and removed in batches on the collection path: the
// delete names no member, so the list id is a parent, not the member id.
// A response field made required names the object.
func TestCollectionOperationAndRequiredName(t *testing.T) {
	r, err := deriveYAML(t, `type: telnyx_list_member
description: x
create: POST /lists/{list_id}/members
delete: DELETE /lists/{list_id}/members
read_from: {op: "GET /lists/{list_id}/members", key: phone_number}
attributes:
  phone_numbers: {skip: true}
  phone_number: {required: true}
`)
	if err != nil {
		t.Fatal(err)
	}
	if l := r.Attr("list_id"); l == nil || !slices.Equal(l.PathParams, []string{"list_id"}) {
		t.Errorf("list_id %+v", l)
	}
	pn := r.Attr("phone_number")
	if r.IDAttr != "phone_number" || pn == nil || !pn.Required || !pn.ForceNew || len(pn.PathParams) != 0 {
		t.Errorf("id %s, phone_number %+v", r.IDAttr, pn)
	}
	if !slices.Equal(r.Import, []string{"list_id", "phone_number"}) {
		t.Errorf("import %v", r.Import)
	}
}

// An extra update without a clear keeps the API's value (computed); an
// attribute create requires stays required.
func TestExtraUpdateWithoutClear(t *testing.T) {
	r, err := deriveYAML(t, `type: telnyx_gadget
description: x
create: POST /gadgets
read: GET /gadgets/{gadget_id}
delete: DELETE /gadgets/{gadget_id}
updates:
  - op: POST /gadgets/{gadget_id}/actions/set_blocklist
`)
	if err != nil {
		t.Fatal(err)
	}
	b := r.Attr("blocklist_id")
	if b == nil || !b.Optional || !b.Computed || b.Required {
		t.Errorf("blocklist_id %+v", b)
	}
}

// A data-source lookup can leave out a query parameter it cannot express.
func TestSkippedQueryLookup(t *testing.T) {
	body := "type: telnyx_lookup\nkind: data_source\ndescription: x\nread: GET /lookups/{number}\n"
	// Boolean query parameters are lookup attributes (wave 5).
	if r, err := deriveYAML(t, body); err != nil || r.Attr("fresh") == nil || r.Attr("fresh").Kind != engine.KindBool {
		t.Errorf("boolean query parameter: %v", err)
	}
	r, err := deriveYAML(t, body+"attributes:\n  fresh: {skip: true}\n")
	if err != nil {
		t.Fatal(err)
	}
	if r.Attr("fresh") != nil || r.Attr("number") == nil {
		t.Errorf("attrs %+v", r.Attrs)
	}
}
