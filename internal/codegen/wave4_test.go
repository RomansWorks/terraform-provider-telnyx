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

const poolOverride = "type: telnyx_pool\ndescription: x\ncreate: POST /pools\nread: \"GET /pools/{pool_id}\"\n" +
	"delete: \"DELETE /pools/{pool_id}\"\nlist: GET /pools\n"

func loadInline(t *testing.T, body string) (*Override, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return LoadOverride(p)
}

// TestWave4Shapes covers force_destroy and lists that answer with a union
// of page shapes.
func TestWave4Shapes(t *testing.T) {
	g := fixture(t)
	ov, err := loadInline(t, poolOverride+"force_destroy: [force, drain, cascade]\n")
	if err != nil {
		t.Fatal(err)
	}
	r, l, err := g.DeriveAll(ov)
	if err != nil {
		t.Fatal(err)
	}
	// cascade is a $ref'd parameter that takes a boolean or a string.
	if strings.Join(r.ForceDestroy, ",") != "force,drain,cascade" {
		t.Errorf("ForceDestroy: %v", r.ForceDestroy)
	}
	if a := r.Attr(engine.ForceDestroyAttr); a == nil || !a.Local || !a.Optional || a.Computed || a.ForceNew || a.Kind != engine.KindBool ||
		!strings.Contains(a.Description, "`force=true` and `drain=true` and `cascade=true`") {
		t.Errorf("force_destroy: %+v", a)
	}
	// GET /pools answers with oneOf a page-number and a cursor shape.
	if l == nil || l.Items == nil || l.Op.Paging != engine.PagingNumber || !l.Op.Envelope {
		t.Fatalf("list: %+v", l)
	}
	if l.Items.Attrs == nil || len(l.Items.Attrs) != 2 {
		t.Errorf("list items: %+v", l.Items.Attrs)
	}

	plain, _, err := g.DeriveAll(mustLoad(t, poolOverride))
	if err != nil {
		t.Fatal(err)
	}
	if plain.ForceDestroy != nil || plain.Attr(engine.ForceDestroyAttr) != nil {
		t.Errorf("force_destroy without the key: %+v", plain.ForceDestroy)
	}
}

func mustLoad(t *testing.T, body string) *Override {
	t.Helper()
	ov, err := loadInline(t, body)
	if err != nil {
		t.Fatal(err)
	}
	return ov
}

func TestWave4Errors(t *testing.T) {
	g := fixture(t)
	cases := map[string]string{
		"force_destroy without delete": "type: telnyx_x\ndescription: x\ncreate: POST /pools\nread: \"GET /pools/{pool_id}\"\nforce_destroy: [force]\n",
		"force_destroy unknown param":  poolOverride + "force_destroy: [purge]\n",
		"force_destroy string param":   poolOverride + "force_destroy: [reason]\n",
		"force_destroy on data source": "type: telnyx_x\nkind: data_source\ndescription: x\nread: \"GET /pools/{pool_id}\"\nforce_destroy: [force]\n",
	}
	for name, body := range cases {
		ov, err := loadInline(t, body)
		if err == nil {
			_, err = g.Derive(ov)
		}
		if err == nil {
			t.Errorf("%s: want an error", name)
		} else {
			t.Logf("%s: %v", name, err)
		}
	}
}
