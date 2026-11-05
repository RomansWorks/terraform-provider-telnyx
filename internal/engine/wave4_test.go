// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// pool is deleted with force=true&drain=true when force_destroy is true.
func pool() *Resource {
	return &Resource{
		TypeName:     "telnyx_pool",
		Create:       &Op{Method: "POST", Path: "/pools", Envelope: true},
		Read:         &Op{Method: "GET", Path: "/pools/{id}", Envelope: true},
		Update:       &Op{Method: "PATCH", Path: "/pools/{id}", Envelope: true},
		Delete:       &Op{Method: "DELETE", Path: "/pools/{id}"},
		ForceDestroy: []string{"force", "drain"},
		IDAttr:       "id",
		Import:       []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: ForceDestroyAttr, Kind: KindBool, Optional: true, Local: true},
		},
	}
}

// deleteQueries records the query string of every DELETE before the fake
// API handles it.
type deleteQueries struct {
	next    http.Handler
	mu      sync.Mutex
	queries []string
}

func (d *deleteQueries) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		d.mu.Lock()
		d.queries = append(d.queries, r.URL.RawQuery)
		d.mu.Unlock()
	}
	d.next.ServeHTTP(w, r)
}

func TestForceDestroy(t *testing.T) {
	for name, tc := range map[string]struct {
		config, want string
	}{
		"not set": {`resource "telnyx_pool" "p" { name = "p" }`, ""},
		"false":   {`resource "telnyx_pool" "p" {` + "\n" + `  name = "p"` + "\n" + `  force_destroy = false` + "\n}", ""},
		"true":    {`resource "telnyx_pool" "p" {` + "\n" + `  name = "p"` + "\n" + `  force_destroy = true` + "\n}", "drain=true&force=true"},
	} {
		t.Run(name, func(t *testing.T) {
			api, _ := newFakeAPI(t)
			rec := &deleteQueries{next: api}
			srv := httptest.NewServer(rec)
			t.Cleanup(srv.Close)
			c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
			if err != nil {
				t.Fatal(err)
			}
			rtest.UnitTest(t, rtest.TestCase{
				ProtoV6ProviderFactories: factories(c, pool()),
				CheckDestroy: func(*terraform.State) error {
					rec.mu.Lock()
					defer rec.mu.Unlock()
					if len(rec.queries) != 1 || rec.queries[0] != tc.want || api.count() != 0 {
						return fmt.Errorf("delete queries %q, want [%q]", rec.queries, tc.want)
					}
					return nil
				},
				Steps: []rtest.TestStep{{Config: tc.config}},
			})
		})
	}
}
