// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// fakeAPI is an in-memory Telnyx-like API: POST /{kind} creates, and
// GET/PATCH/PUT/DELETE /{kind}/{id} act on one object.
type fakeAPI struct {
	mu       sync.Mutex
	objects  map[string]Object // by path
	next     int
	requests []string // "METHOD path body"
	envelope bool
}

func newFakeAPI(t *testing.T) (*fakeAPI, *client.Client) {
	t.Helper()
	f := &fakeAPI{objects: map[string]Object{}, envelope: true}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	return f, c
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, strings.TrimSpace(r.Method+" "+r.URL.EscapedPath()+" "+string(body)))
	var in Object
	if len(body) > 0 {
		if err := json.Unmarshal(body, &in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	p := r.URL.EscapedPath()
	segs := strings.Split(strings.Trim(p, "/"), "/")
	switch {
	case r.Method == http.MethodPost && len(segs) == 1:
		f.next++
		id := fmt.Sprintf("%s-%d", segs[0], f.next)
		in["id"] = id
		in["created_at"] = "2026-10-29T00:00:00Z"
		in["status"] = "pending"
		f.objects[p+"/"+id] = in
		f.write(w, http.StatusCreated, in)
		return
	}
	obj, ok := f.objects[p]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"code":"10005","title":"Resource not found"}]}`))
		return
	}
	switch r.Method {
	case http.MethodGet:
		if obj["status"] == "pending" {
			obj["status"] = "ready" // ready on the second look
		}
	case http.MethodPatch:
		for k, v := range in {
			if v == nil {
				delete(obj, k)
			} else {
				obj[k] = v
			}
		}
	case http.MethodPut:
		for k := range obj {
			if k != "id" && k != "created_at" && k != "status" {
				delete(obj, k)
			}
		}
		for k, v := range in {
			obj[k] = v
		}
	case http.MethodDelete:
		delete(f.objects, p)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	f.write(w, http.StatusOK, obj)
}

func (f *fakeAPI) write(w http.ResponseWriter, status int, obj Object) {
	var out any = obj
	if f.envelope {
		out = map[string]any{"data": obj}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(out)
}

func (f *fakeAPI) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

func (f *fakeAPI) has(prefix string) bool {
	for _, r := range f.log() {
		if strings.HasPrefix(r, prefix) {
			return true
		}
	}
	return false
}

func (f *fakeAPI) remove(path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.objects, path)
}

func (f *fakeAPI) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.objects)
}

// widget has every lifecycle operation, a data envelope, PATCH updates and a
// wait hook.
func widget() *Resource {
	return &Resource{
		TypeName: "telnyx_widget",
		Create:   &Op{Method: "POST", Path: "/widgets", Envelope: true},
		Read:     &Op{Method: "GET", Path: "/widgets/{id}", Envelope: true},
		Update:   &Op{Method: "PATCH", Path: "/widgets/{id}", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/widgets/{id}"},
		IDAttr:   "id",
		Import:   []string{"id"},
		Timeouts: true,
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParam: "id"},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "description", JSON: "description", Kind: KindString, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "settings", JSON: "settings", Kind: KindJSON, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "tags", JSON: "tags", Kind: KindSet, Elem: &Attr{Kind: KindString}, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "region", JSON: "region", Kind: KindString, Optional: true, ForceNew: true, InCreate: true, InResponse: true, Enum: []string{"eu", "us"}},
			{Name: "status", JSON: "status", Kind: KindString, Computed: true, InResponse: true},
			{Name: "created_at", JSON: "created_at", Kind: KindString, Computed: true, InResponse: true},
		},
		DataSource: true,
		Hooks: Hooks{
			Wait: func(ctx context.Context, c *client.Client, r *Resource, obj Object) (Object, error) {
				for obj["status"] != "ready" {
					var err error
					obj, err = Get(ctx, c, r, Values{"id": str(obj["id"].(string))}) //nolint:forcetypeassert
					if err != nil {
						return nil, err
					}
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(10 * time.Millisecond):
					}
				}
				return obj, nil
			},
		},
	}
}

// gadget cannot be deleted, is replaced with PUT and has no envelope.
func gadget() *Resource {
	return &Resource{
		TypeName: "telnyx_gadget",
		Create:   &Op{Method: "POST", Path: "/gadgets", Envelope: true},
		Read:     &Op{Method: "GET", Path: "/gadgets/{gadget_id}", Envelope: true},
		Update:   &Op{Method: "PUT", Path: "/gadgets/{gadget_id}", Envelope: true},
		IDAttr:   "id",
		Import:   []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParam: "gadget_id"},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "size", JSON: "size", Kind: KindInt, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "api_secret", JSON: "secret", Kind: KindString, Optional: true, Sensitive: true, InCreate: true, InUpdate: true},
			{Name: "mode_a", JSON: "mode_a", Kind: KindString, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "mode_b", JSON: "mode_b", Kind: KindString, Optional: true, InCreate: true, InUpdate: true, InResponse: true},
		},
		ExactlyOne: [][]string{{"mode_a", "mode_b"}},
	}
}

type testProvider struct {
	c   *client.Client
	res []*Resource
}

func (p *testProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "telnyx"
}

func (p *testProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {}

func (p *testProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.ResourceData = p.c
	resp.DataSourceData = p.c
}

func (p *testProvider) Resources(context.Context) []func() resource.Resource {
	var out []func() resource.Resource
	for _, r := range p.res {
		out = append(out, NewResource(r))
	}
	return out
}

func (p *testProvider) DataSources(context.Context) []func() datasource.DataSource {
	var out []func() datasource.DataSource
	for _, r := range p.res {
		if r.DataSource {
			out = append(out, NewDataSource(r))
		}
	}
	return out
}

func factories(c *client.Client, res ...*Resource) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"telnyx": providerserver.NewProtocol6WithError(&testProvider{c: c, res: res}),
	}
}

func TestSchemasAreValid(t *testing.T) {
	ctx := context.Background()
	for _, d := range []*Resource{widget(), gadget()} {
		if diags := d.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s resource schema: %v", d.TypeName, diags)
		}
		if diags := d.DataSourceSchema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s data source schema: %v", d.TypeName, diags)
		}
	}
}

func TestLifecyclePatch(t *testing.T) {
	api, c := newFakeAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, widget()),
		CheckDestroy: func(*terraform.State) error {
			if api.count() != 0 {
				return fmt.Errorf("%d objects left", api.count())
			}
			return nil
		},
		Steps: []rtest.TestStep{
			{
				Config: `
resource "telnyx_widget" "w" {
  name        = "one"
  description = "first"
  settings    = jsonencode({ b = 1, a = [true] })
  tags        = ["x", "y"]
  region      = "eu"
}
data "telnyx_widget" "w" {
  id = telnyx_widget.w.id
}`,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("telnyx_widget.w", "id", "widgets-1"),
					rtest.TestCheckResourceAttr("telnyx_widget.w", "status", "ready"),
					rtest.TestCheckResourceAttr("telnyx_widget.w", "created_at", "2026-10-29T00:00:00Z"),
					rtest.TestCheckResourceAttr("telnyx_widget.w", "settings", `{"a":[true],"b":1}`),
					rtest.TestCheckResourceAttr("data.telnyx_widget.w", "name", "one"),
					rtest.TestCheckResourceAttr("data.telnyx_widget.w", "tags.#", "2"),
					rtest.TestCheckResourceAttr("data.telnyx_widget.w", "settings", `{"a":[true],"b":1}`),
					func(*terraform.State) error {
						if !api.has(`POST /widgets {"description":"first","name":"one","region":"eu","settings":{"a":[true],"b":1},"tags":["x","y"]}`) {
							return fmt.Errorf("unexpected create request: %v", api.log())
						}
						return nil
					},
				),
			},
			{
				ResourceName:            "telnyx_widget.w",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				// Rename, drop the description: only changed fields are sent,
				// and the removed one is cleared with null.
				Config: `
resource "telnyx_widget" "w" {
  name     = "two"
  settings = jsonencode({ a = [true], b = 1 })
  tags     = ["y", "x"]
  region   = "eu"
}`,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_widget.w", plancheck.ResourceActionUpdate)},
				},
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("telnyx_widget.w", "name", "two"),
					rtest.TestCheckNoResourceAttr("telnyx_widget.w", "description"),
					func(*terraform.State) error {
						if !api.has(`PATCH /widgets/widgets-1 {"description":null,"name":"two"}`) {
							return fmt.Errorf("unexpected update request: %v", api.log())
						}
						return nil
					},
				),
			},
			{
				// A ForceNew attribute replaces the object.
				Config: `
resource "telnyx_widget" "w" {
  name     = "two"
  settings = jsonencode({ a = [true], b = 1 })
  tags     = ["y", "x"]
  region   = "us"
}`,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_widget.w", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: rtest.TestCheckResourceAttr("telnyx_widget.w", "id", "widgets-2"),
			},
			{
				// Deleted outside Terraform: read drops it and the plan
				// recreates it.
				PreConfig: func() { api.remove("/widgets/widgets-2") },
				Config: `
resource "telnyx_widget" "w" {
  name     = "two"
  settings = jsonencode({ a = [true], b = 1 })
  tags     = ["y", "x"]
  region   = "us"
}`,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_widget.w", plancheck.ResourceActionCreate)},
				},
				Check: rtest.TestCheckResourceAttr("telnyx_widget.w", "id", "widgets-3"),
			},
		},
	})
}

func TestLifecyclePutWithoutDelete(t *testing.T) {
	api, c := newFakeAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, gadget()),
		CheckDestroy: func(*terraform.State) error {
			if api.count() != 1 {
				return fmt.Errorf("the object must stay in the API, found %d", api.count())
			}
			return nil
		},
		Steps: []rtest.TestStep{
			{
				Config:      `resource "telnyx_gadget" "g" { name = "g" }`,
				ExpectError: regexpMust(`Exactly one of these attributes must be configured`),
			},
			{
				Config: `
resource "telnyx_gadget" "g" {
  name       = "g"
  size       = 2
  api_secret = "s"
  mode_a     = "on"
}`,
				Check: rtest.TestCheckResourceAttr("telnyx_gadget.g", "api_secret", "s"),
			},
			{
				// PUT sends every attribute, changed or not.
				Config: `
resource "telnyx_gadget" "g" {
  name       = "g"
  size       = 3
  api_secret = "s"
  mode_a     = "on"
}`,
				Check: func(*terraform.State) error {
					if !api.has(`PUT /gadgets/gadgets-1 {"mode_a":"on","name":"g","secret":"s","size":3}`) {
						return fmt.Errorf("unexpected update request: %v", api.log())
					}
					return nil
				},
			},
			{
				ResourceName:            "telnyx_gadget.g",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_secret"},
			},
		},
	})
	if api.has("DELETE") {
		t.Error("no delete request may be sent")
	}
}

func regexpMust(s string) *regexp.Regexp { return regexp.MustCompile(s) }
