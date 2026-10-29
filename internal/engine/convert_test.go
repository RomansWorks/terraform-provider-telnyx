// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package engine

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func testAttrs() []*Attr {
	return []*Attr{
		{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParam: "id"},
		{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
		{Name: "count", JSON: "count", Kind: KindInt, Optional: true, InCreate: true, InResponse: true},
		{Name: "ratio", JSON: "ratio", Kind: KindFloat, Optional: true, InCreate: true, InResponse: true},
		{Name: "enabled", JSON: "enabled", Kind: KindBool, Optional: true, Computed: true, InCreate: true, InResponse: true},
		{Name: "settings", JSON: "settings", Kind: KindJSON, Optional: true, InCreate: true, InResponse: true},
		{Name: "tags", JSON: "tags", Kind: KindList, Elem: &Attr{Kind: KindString}, Optional: true, InCreate: true, InResponse: true},
		{Name: "secret", JSON: "secret", Kind: KindString, Optional: true, Sensitive: true, InCreate: true},
		{Name: "file_path", Kind: KindString, Optional: true, Local: true},
		{Name: "owner", JSON: "owner", Kind: KindObject, Optional: true, Computed: true, InCreate: true, InResponse: true, Attrs: []*Attr{
			{Name: "email", JSON: "email", Kind: KindString, Optional: true, InResponse: true},
			{Name: "rank", JSON: "rank", Kind: KindInt, Computed: true, InResponse: true},
		}},
	}
}

func str(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

func num(f float64) tftypes.Value { return tftypes.NewValue(tftypes.Number, big.NewFloat(f)) }

func nullOf(a []*Attr) Values {
	out := Values{}
	for _, x := range a {
		out[x.Name] = tftypes.NewValue(x.TerraformType(), nil)
	}
	return out
}

func decode(t *testing.T, s string) Object {
	t.Helper()
	obj, err := decodeObject([]byte(s), false)
	if err != nil {
		t.Fatal(err)
	}
	return obj
}

func TestObjectToJSON(t *testing.T) {
	attrs := testAttrs()
	v := nullOf(attrs)
	v["name"] = str("a")
	v["count"] = num(3)
	v["ratio"] = num(0.5)
	v["enabled"] = tftypes.NewValue(tftypes.Bool, false)
	v["settings"] = str(`{"b": 1, "a": [true]}`)
	v["tags"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{str("x"), str("y")})
	v["file_path"] = str("/tmp/x")
	v["id"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

	obj, err := objectToJSON(attrs, v, func(a *Attr) bool { return a.InCreate })
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(obj)
	want := `{"count":3,"enabled":false,"name":"a","ratio":0.5,"settings":{"b":1,"a":[true]},"tags":["x","y"]}`
	if string(data) != want {
		t.Errorf("got  %s\nwant %s", data, want)
	}
}

func TestObjectToJSONInvalidJSON(t *testing.T) {
	attrs := testAttrs()
	v := nullOf(attrs)
	v["settings"] = str(`{`)
	if _, err := objectToJSON(attrs, v, func(*Attr) bool { return true }); err == nil {
		t.Fatal("want an error for invalid JSON")
	}
}

func TestObjectFromJSONApplyKeepsPlan(t *testing.T) {
	attrs := testAttrs()
	plan := nullOf(attrs)
	plan["id"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	plan["name"] = str("a")
	plan["settings"] = str(`{ "a": 1 }`)
	plan["enabled"] = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	plan["owner"] = tftypes.NewValue(attrs[9].TerraformType(), tftypes.UnknownValue)
	plan["secret"] = str("s3cret")

	obj := decode(t, `{"id":"42","name":"A","settings":{"a":1},"enabled":true,"count":7,"owner":{"email":"e","rank":2}}`)
	v, err := objectFromJSON(attrs, obj, plan, modeApply)
	if err != nil {
		t.Fatal(err)
	}
	f := fields(v)
	check := func(name string, want tftypes.Value) {
		t.Helper()
		if !f[name].Equal(want) {
			t.Errorf("%s = %s, want %s", name, f[name], want)
		}
	}
	check("id", str("42"))
	check("name", str("a"))              // configured value kept
	check("settings", str(`{ "a": 1 }`)) // user formatting kept
	check("enabled", tftypes.NewValue(tftypes.Bool, true))
	check("count", tftypes.NewValue(tftypes.Number, nil)) // not configured, not computed
	check("secret", str("s3cret"))                        // not in the response: kept
	var owner map[string]tftypes.Value
	if err := f["owner"].As(&owner); err != nil {
		t.Fatal(err)
	}
	if !owner["rank"].Equal(num(2)) || !owner["email"].Equal(str("e")) {
		t.Errorf("owner = %v", owner)
	}
	if !v.IsFullyKnown() {
		t.Error("state must be fully known")
	}
}

func TestObjectFromJSONReadDetectsDrift(t *testing.T) {
	attrs := testAttrs()
	state := nullOf(attrs)
	state["id"] = str("42")
	state["name"] = str("a")
	state["count"] = num(3)
	state["settings"] = str(`{"a":1}`)
	state["file_path"] = str("/tmp/x")

	obj := decode(t, `{"id":"42","name":"b","settings":{"a":2}}`)
	v, err := objectFromJSON(attrs, obj, state, modeRead)
	if err != nil {
		t.Fatal(err)
	}
	f := fields(v)
	if !f["name"].Equal(str("b")) {
		t.Errorf("name = %s, want drift to b", f["name"])
	}
	if !f["settings"].Equal(str(`{"a":2}`)) {
		t.Errorf("settings = %s", f["settings"])
	}
	if !f["count"].Equal(num(3)) {
		t.Errorf("count = %s, want prior kept when absent", f["count"])
	}
	if !f["file_path"].Equal(str("/tmp/x")) {
		t.Errorf("local attribute must be kept, got %s", f["file_path"])
	}
}

func TestObjectFromJSONNullClears(t *testing.T) {
	attrs := testAttrs()
	state := nullOf(attrs)
	state["count"] = num(3)
	v, err := objectFromJSON(attrs, decode(t, `{"count":null}`), state, modeRead)
	if err != nil {
		t.Fatal(err)
	}
	if !fields(v)["count"].IsNull() {
		t.Error("an explicit null from the API must clear the value")
	}
}

func TestFromJSONTypeErrors(t *testing.T) {
	attrs := testAttrs()
	for _, body := range []string{`{"name":{}}`, `{"count":"x"}`, `{"enabled":"yes"}`, `{"tags":"x"}`, `{"owner":[]}`} {
		if _, err := objectFromJSON(attrs, decode(t, body), nil, modeData); err == nil {
			t.Errorf("%s: want an error", body)
		}
	}
}

func TestFromJSONCoercesScalarsToString(t *testing.T) {
	a := &Attr{Kind: KindString}
	for raw, want := range map[any]string{json.Number("12"): "12", true: "true"} {
		v, err := fromJSON(a, raw, tftypes.Value{}, modeData)
		if err != nil || !v.Equal(str(want)) {
			t.Errorf("%v: got %s, %v", raw, v, err)
		}
	}
}

func TestLargeIntegersArePrecise(t *testing.T) {
	a := &Attr{Kind: KindInt}
	v, err := fromJSON(a, json.Number("9007199254740993"), tftypes.Value{}, modeData)
	if err != nil {
		t.Fatal(err)
	}
	j, _, err := toJSON(a, v)
	if err != nil || j != json.Number("9007199254740993") {
		t.Errorf("round trip = %v, %v", j, err)
	}
}

func TestKnown(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String, "l": tftypes.List{ElementType: tftypes.String}}}
	v := tftypes.NewValue(typ, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"l": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}),
	})
	if !known(v).IsFullyKnown() {
		t.Error("known must remove every unknown")
	}
}

func TestJSONEqual(t *testing.T) {
	if !JSONEqual(`{"a":1,"b":[1,2]}`, "{\n \"b\": [1, 2], \"a\": 1}") {
		t.Error("same content must be equal")
	}
	if JSONEqual(`{"a":1}`, `{"a":1.5}`) || JSONEqual(`{`, `{`) {
		t.Error("different or invalid documents must not be equal")
	}
}

func TestDecodeObjectEnvelope(t *testing.T) {
	obj, err := decodeObject([]byte(`{"data":{"id":"1"}}`), true)
	if err != nil || obj["id"] != "1" {
		t.Errorf("got %v, %v", obj, err)
	}
	if _, err := decodeObject([]byte(`{"id":"1"}`), true); err == nil {
		t.Error("missing envelope must be an error")
	}
	if obj, err := decodeObject(nil, true); err != nil || len(obj) != 0 {
		t.Errorf("empty body: %v, %v", obj, err)
	}
}

func TestFillPath(t *testing.T) {
	d := &Resource{Attrs: testAttrs()}
	p, err := fillPath(d, "/things/{id}/voice", Values{"id": str("a/b c")})
	if err != nil || p != "/things/a%2Fb%20c/voice" {
		t.Errorf("got %q, %v", p, err)
	}
	if _, err := fillPath(d, "/things/{id}", Values{"id": tftypes.NewValue(tftypes.String, nil)}); err == nil {
		t.Error("a null id must be an error")
	}
	if _, err := fillPath(d, "/things/{other}", Values{}); err == nil {
		t.Error("an unknown path parameter must be an error")
	}
}
