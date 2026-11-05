// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

// TestForceDestroyWithoutUpdate sets force_destroy on a resource that has
// no update operation: the change is stored without a request.
func TestForceDestroyWithoutUpdate(t *testing.T) {
	api, c := newFakeAPI(t)
	d := pool()
	d.Update = nil
	d.Attrs[1].InUpdate = false
	d.Attrs[1].ForceNew = true
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, d),
		Steps: []rtest.TestStep{
			{Config: `resource "telnyx_pool" "p" { name = "p" }`},
			{
				Config: "resource \"telnyx_pool\" \"p\" {\n  name = \"p\"\n  force_destroy = true\n}",
				ConfigPlanChecks: rtest.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("telnyx_pool.p", plancheck.ResourceActionUpdate),
				}},
				Check: func(*terraform.State) error {
					if api.has("PATCH") || api.has("PUT") {
						return fmt.Errorf("no update request expected: %v", api.log())
					}
					return nil
				},
			},
		},
	})
}

// TestReadGone removes an object from state when the read hook reports it
// soft-deleted with a not-found error.
func TestReadGone(t *testing.T) {
	_, c := newFakeAPI(t)
	d := pool()
	gone := false
	d.Hooks.AfterRead = func(_ context.Context, obj Object) (Object, error) {
		if gone {
			return nil, &client.Error{StatusCode: http.StatusNotFound}
		}
		return obj, nil
	}
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, d),
		Steps: []rtest.TestStep{
			{Config: `resource "telnyx_pool" "p" { name = "p" }`},
			{
				PreConfig:          func() { gone = true },
				Config:             `resource "telnyx_pool" "p" { name = "p" }`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{PreConfig: func() { gone = false }, Config: `resource "telnyx_pool" "p" { name = "p" }`},
		},
	})
}

func TestLowercase(t *testing.T) {
	d := pool()
	d.Attrs[1].Lowercase = true
	_, c := newFakeAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, d),
		Steps: []rtest.TestStep{
			{Config: `resource "telnyx_pool" "p" { name = "Mixed@Example.com" }`, ExpectError: regexp.MustCompile(`must be lowercase`)},
			{Config: `resource "telnyx_pool" "p" { name = " a@example.com" }`, ExpectError: regexp.MustCompile(`must be lowercase`)},
			{Config: `resource "telnyx_pool" "p" { name = "a.b+c@example.com" }`},
		},
	})
}
