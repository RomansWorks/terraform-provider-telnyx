// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestCreateOnly keeps a value only create returned, whatever read says,
// and takes it from the create response.
func TestCreateOnly(t *testing.T) {
	attrs := []*Attr{{Name: "token", JSON: "token", Kind: KindString, Computed: true, InResponse: true, CreateOnly: true}}
	prior := map[string]tftypes.Value{"token": tftypes.NewValue(tftypes.String, "t1")}
	for _, obj := range []Object{{}, {"token": nil}, {"token": "other"}} {
		v, err := objectFromJSON(attrs, obj, prior, modeRead)
		if err != nil {
			t.Fatal(err)
		}
		var got string
		_ = fields(v)["token"].As(&got)
		if got != "t1" {
			t.Errorf("read %v: got %q", obj, got)
		}
	}
	unknown := map[string]tftypes.Value{"token": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}
	v, err := objectFromJSON(attrs, Object{"token": "t2"}, unknown, modeApply)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	_ = fields(v)["token"].As(&got)
	if got != "t2" {
		t.Errorf("create: got %q", got)
	}
}
