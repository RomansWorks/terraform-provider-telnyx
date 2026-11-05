// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// TestCreateOnlyRead keeps a value only create returned, whatever read says.
func TestCreateOnlyRead(t *testing.T) {
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
}

// TestCreateOnlyLifecycle: the token comes only from the create response;
// the read after create, refreshes and updates keep it.
func TestCreateOnlyLifecycle(t *testing.T) {
	var mu sync.Mutex
	vault := Object{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var in Object
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &in)
		var out Object
		switch r.Method {
		case http.MethodPost:
			vault = Object{"id": "v1", "name": in["name"]}
			out = Object{"id": "v1", "name": in["name"], "token": "t1"}
		case http.MethodPatch:
			vault["name"] = in["name"]
			out = Object{"id": "v1", "name": vault["name"]}
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			out = Object{"id": "v1", "name": vault["name"], "token": nil}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	d := &Resource{
		TypeName: "telnyx_vault",
		Create:   &Op{Method: "POST", Path: "/vaults", Envelope: true},
		Read:     &Op{Method: "GET", Path: "/vaults/{id}", Envelope: true},
		Update:   &Op{Method: "PATCH", Path: "/vaults/{id}", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/vaults/{id}"},
		IDAttr:   "id",
		Import:   []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "token", JSON: "token", Kind: KindString, Computed: true, Sensitive: true, InResponse: true, CreateOnly: true},
		},
	}
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, d),
		Steps: []rtest.TestStep{
			{
				Config: `resource "telnyx_vault" "v" { name = "one" }`,
				Check:  rtest.TestCheckResourceAttr("telnyx_vault.v", "token", "t1"),
			},
			{
				Config: `resource "telnyx_vault" "v" { name = "two" }`,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("telnyx_vault.v", "name", "two"),
					rtest.TestCheckResourceAttr("telnyx_vault.v", "token", "t1"),
				),
			},
		},
	})
}
