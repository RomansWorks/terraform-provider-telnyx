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
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// collectionAPI serves collections without item reads:
//   - /things: a read-only list paged with page[number] and meta.total_pages
//   - /regions: a list of plain values with no paging
//   - /secrets: create, list (page_number, no total) and delete
//   - /boxes/box-1: an object with tags and extra update actions
type collectionAPI struct {
	mu        sync.Mutex
	things    []Object
	secrets   []Object
	box       Object
	next      int
	requests  []string
	withTotal bool
}

func newCollectionAPI(t *testing.T) (*collectionAPI, *client.Client) {
	t.Helper()
	a := &collectionAPI{withTotal: true, box: Object{"id": "box-1", "name": "b", "tags": []any{}}}
	for i := 1; i <= 7; i++ {
		name := "us"
		if i%2 == 1 {
			name = "eu"
		}
		a.things = append(a.things, Object{"id": fmt.Sprintf("t%d", i), "name": name})
	}
	srv := httptest.NewServer(a)
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	return a, c
}

func (a *collectionAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	body, _ := io.ReadAll(r.Body)
	q := r.URL.Query()
	a.requests = append(a.requests, strings.TrimSpace(r.Method+" "+r.URL.Path+"?"+q.Encode()+" "+string(body)))
	var in Object
	_ = json.Unmarshal(body, &in)
	reply := func(v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	page := func(all []Object, number, size string) {
		n, _ := strconv.Atoi(number)
		s, _ := strconv.Atoi(size)
		if n < 1 {
			n = 1
		}
		if s < 1 {
			s = 2
		}
		lo, hi := min((n-1)*s, len(all)), min(n*s, len(all))
		out := map[string]any{"data": all[lo:hi]}
		if a.withTotal {
			out["meta"] = map[string]any{"total_pages": (len(all) + s - 1) / s}
		}
		reply(out)
	}
	notFound := func() { http.Error(w, `{"errors":[{"code":"10005"}]}`, http.StatusNotFound) }

	switch p := r.URL.Path; {
	case r.Method == http.MethodGet && p == "/things":
		var match []Object
		for _, t := range a.things {
			if f := q.Get("filter[name]"); f == "" || t["name"] == f {
				match = append(match, t)
			}
		}
		page(match, q.Get("page[number]"), q.Get("page[size]"))
	case r.Method == http.MethodGet && p == "/regions":
		reply(map[string]any{"data": []any{"eu", "us"}})
	case r.Method == http.MethodPost && p == "/secrets":
		a.next++
		obj := Object{"id": fmt.Sprintf("s%d", a.next), "name": in["name"]}
		a.secrets = append(a.secrets, obj)
		reply(map[string]any{"data": obj})
	case r.Method == http.MethodGet && p == "/secrets":
		page(a.secrets, q.Get("page_number"), q.Get("page_size"))
	case r.Method == http.MethodDelete && strings.HasPrefix(p, "/secrets/"):
		id := strings.TrimPrefix(p, "/secrets/")
		a.secrets = slices.DeleteFunc(a.secrets, func(o Object) bool { return o["id"] == id })
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPost && p == "/boxes":
		a.box = Object{"id": "box-1", "name": in["name"], "tags": []any{}}
		reply(map[string]any{"data": a.box})
	case strings.HasPrefix(p, "/boxes/box-1") && a.box == nil:
		notFound()
	case r.Method == http.MethodGet && p == "/boxes/box-1":
		reply(map[string]any{"data": a.box})
	case r.Method == http.MethodPatch && p == "/boxes/box-1":
		for k, v := range in {
			a.box[k] = v
		}
		reply(map[string]any{"data": a.box})
	case r.Method == http.MethodDelete && p == "/boxes/box-1":
		a.box = nil
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPost && p == "/boxes/box-1/actions/set_blocklist":
		a.box["blocklist_id"] = in["blocklist_id"]
		reply(map[string]any{"data": a.box})
	case r.Method == http.MethodPost && p == "/boxes/box-1/actions/remove_blocklist":
		delete(a.box, "blocklist_id")
		reply(map[string]any{"data": a.box})
	case r.Method == http.MethodPost && p == "/boxes/box-1/tags":
		a.box["tags"] = append(a.box["tags"].([]any), in["tag"]) //nolint:forcetypeassert
		reply(map[string]any{"data": a.box})
	case r.Method == http.MethodDelete && strings.HasPrefix(p, "/boxes/box-1/tags/"):
		tag := strings.TrimPrefix(p, "/boxes/box-1/tags/")
		a.box["tags"] = slices.DeleteFunc(a.box["tags"].([]any), func(v any) bool { return v == tag }) //nolint:forcetypeassert
		reply(map[string]any{"data": a.box})
	default:
		notFound()
	}
}

func (a *collectionAPI) log() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.requests...)
}

// sent counts the requests that start with prefix.
func (a *collectionAPI) sent(prefix string) int {
	n := 0
	for _, r := range a.log() {
		if strings.HasPrefix(r, prefix) {
			n++
		}
	}
	return n
}

func (a *collectionAPI) set(fn func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fn()
}

func thingsList() *List {
	return &List{
		TypeName: "telnyx_things",
		Op:       &Op{Method: "GET", Path: "/things", Envelope: true, Paging: PagingNumber, PageSize: 2},
		Params: []*Attr{
			{Name: "filter_name", JSON: "filter[name]", Kind: KindString, Optional: true},
		},
		Items: &Attr{Name: "things", JSON: "data", Kind: KindList, Computed: true, InResponse: true, Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true},
			{Name: "name", JSON: "name", Kind: KindString, Computed: true, InResponse: true},
		}},
	}
}

func regionsList() *List {
	return &List{
		TypeName: "telnyx_regions",
		Op:       &Op{Method: "GET", Path: "/regions", Envelope: true},
		Params:   []*Attr{},
		Items:    &Attr{Name: "regions", JSON: "data", Kind: KindList, Computed: true, InResponse: true, Elem: &Attr{Kind: KindString}},
	}
}

// secret is read by finding it in its list.
func secret() *Resource {
	return &Resource{
		TypeName: "telnyx_secret",
		Create:   &Op{Method: "POST", Path: "/secrets", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/secrets/{id}"},
		ReadFrom: &ReadFrom{Op: &Op{Method: "GET", Path: "/secrets", Envelope: true, Paging: PagingFlat, PageSize: 2}, Key: "id"},
		IDAttr:   "id",
		Import:   []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, ForceNew: true, InCreate: true, InResponse: true},
			{Name: "value", JSON: "value", Kind: KindString, Required: true, Sensitive: true, ForceNew: true, InCreate: true},
		},
	}
}

// boxTag is a plain value in its parent's tags.
func boxTag() *Resource {
	return &Resource{
		TypeName: "telnyx_box_tag",
		Create:   &Op{Method: "POST", Path: "/boxes/{box_id}/tags", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/boxes/{box_id}/tags/{tag}", Envelope: true},
		ReadFrom: &ReadFrom{Op: &Op{Method: "GET", Path: "/boxes/{box_id}", Envelope: true}, Field: "tags"},
		IDAttr:   "tag",
		Import:   []string{"box_id", "tag"},
		Attrs: []*Attr{
			{Name: "box_id", Kind: KindString, Required: true, ForceNew: true, Local: true, PathParams: []string{"box_id"}},
			{Name: "tag", JSON: "tag", Kind: KindString, Required: true, ForceNew: true, InCreate: true, InResponse: true, PathParams: []string{"tag"}},
		},
	}
}

// box has a blocklist set and removed by their own actions.
func box() *Resource {
	return &Resource{
		TypeName: "telnyx_box",
		Create:   &Op{Method: "POST", Path: "/boxes", Envelope: true},
		Read:     &Op{Method: "GET", Path: "/boxes/{box_id}", Envelope: true},
		Update:   &Op{Method: "PATCH", Path: "/boxes/{box_id}", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/boxes/{box_id}"},
		Updates: []ExtraUpdate{{
			Op:    &Op{Method: "POST", Path: "/boxes/{box_id}/actions/set_blocklist", Envelope: true},
			Clear: &Op{Method: "POST", Path: "/boxes/{box_id}/actions/remove_blocklist"},
			Attrs: []string{"blocklist_id"},
		}},
		IDAttr: "id",
		Import: []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"box_id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "blocklist_id", JSON: "blocklist_id", Kind: KindString, Optional: true, InResponse: true},
		},
	}
}

func TestEachPaging(t *testing.T) {
	api, c := newCollectionAPI(t)
	ctx := context.Background()
	var ids []string
	collect := func(o Object) bool { ids = append(ids, o["id"].(string)); return true } //nolint:forcetypeassert

	if err := Each(ctx, c, thingsList().Op, "/things", nil, collect); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "t1,t2,t3,t4,t5,t6,t7" || api.sent("GET /things") != 4 {
		t.Errorf("number paging: ids %v after %v", ids, api.log())
	}
	if !slices.Contains(api.log(), "GET /things?page%5Bnumber%5D=4&page%5Bsize%5D=2") {
		t.Errorf("page parameters: %v", api.log())
	}

	// Without total_pages, a short page is the last one.
	api.set(func() {
		api.withTotal = false
		api.secrets = []Object{{"id": "a"}, {"id": "b"}, {"id": "c"}}
	})
	ids = nil
	if err := Each(ctx, c, secret().ReadFrom.Op, "/secrets", nil, collect); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "a,b,c" || api.sent("GET /secrets?page_number=") != 2 {
		t.Errorf("flat paging: ids %v after %v", ids, api.log())
	}

	// fn returning false stops before the next page.
	ids = nil
	before := api.sent("GET /things")
	err := Each(ctx, c, thingsList().Op, "/things", url.Values{"filter[name]": {"eu"}}, func(o Object) bool {
		ids = append(ids, o["id"].(string)) //nolint:forcetypeassert
		return len(ids) < 2
	})
	if err != nil || strings.Join(ids, ",") != "t1,t3" || api.sent("GET /things")-before != 1 {
		t.Errorf("early stop: ids %v, err %v, log %v", ids, err, api.log())
	}
}

func TestListDataSource(t *testing.T) {
	api, c := newCollectionAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: listFactories(c, []*List{thingsList(), regionsList()}),
		Steps: []rtest.TestStep{
			{
				Config: `
data "telnyx_things" "all" {}
data "telnyx_things" "eu" {
  filter_name = "eu"
  max_items   = 3
}
data "telnyx_regions" "all" {}`,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("data.telnyx_things.all", "things.#", "7"),
					rtest.TestCheckResourceAttr("data.telnyx_things.all", "things.6.id", "t7"),
					rtest.TestCheckResourceAttr("data.telnyx_things.eu", "things.#", "3"),
					rtest.TestCheckResourceAttr("data.telnyx_things.eu", "things.2.id", "t5"),
					rtest.TestCheckResourceAttr("data.telnyx_regions.all", "regions.#", "2"),
					rtest.TestCheckResourceAttr("data.telnyx_regions.all", "regions.1", "us"),
					func(*terraform.State) error {
						for _, r := range api.log() {
							// Three eu items fill two pages of two; a third page is never read.
							if strings.Contains(r, "filter%5Bname%5D=eu") && strings.Contains(r, "page%5Bnumber%5D=3") {
								return fmt.Errorf("read past max_items: %v", api.log())
							}
						}
						if api.sent("GET /things?filter%5Bname%5D=eu&page%5Bnumber%5D=1") == 0 {
							return fmt.Errorf("filter not sent: %v", api.log())
						}
						return nil
					},
				),
			},
		},
	})
}

func TestReadFromList(t *testing.T) {
	api, c := newCollectionAPI(t)
	api.set(func() {
		api.withTotal = false
		api.secrets = []Object{{"id": "x1", "name": "other"}, {"id": "x2", "name": "other"}}
		api.next = 2
	})
	cfg := `
resource "telnyx_secret" "s" {
  name  = "db"
  value = "hunter2"
}`
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, secret()),
		CheckDestroy: func(*terraform.State) error {
			if api.sent("DELETE /secrets/") == 0 {
				return fmt.Errorf("not deleted: %v", api.log())
			}
			return nil
		},
		Steps: []rtest.TestStep{
			{
				// The new secret is on page 2.
				Config: cfg,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("telnyx_secret.s", "id", "s3"),
					rtest.TestCheckResourceAttr("telnyx_secret.s", "name", "db"),
					func(*terraform.State) error {
						if api.sent("GET /secrets?page_number=2&page_size=2") == 0 {
							return fmt.Errorf("second page not read: %v", api.log())
						}
						return nil
					},
				),
			},
			{
				ResourceName:            "telnyx_secret.s",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			{
				// Gone from the list: removed from state and created again.
				PreConfig: func() {
					api.set(func() {
						api.secrets = slices.DeleteFunc(api.secrets, func(o Object) bool { return o["id"] == "s3" })
					})
				},
				Config: cfg,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_secret.s", plancheck.ResourceActionCreate)},
				},
				Check: rtest.TestCheckResourceAttr("telnyx_secret.s", "id", "s4"),
			},
		},
	})
}

func TestReadFromParentPlainValue(t *testing.T) {
	api, c := newCollectionAPI(t)
	cfg := `
resource "telnyx_box_tag" "vip" {
  box_id = "box-1"
  tag    = "vip"
}`
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, boxTag()),
		Steps: []rtest.TestStep{
			{
				Config: cfg,
				Check: func(*terraform.State) error {
					if api.sent(`POST /boxes/box-1/tags? {"tag":"vip"}`) != 1 {
						return fmt.Errorf("unexpected requests: %v", api.log())
					}
					return nil
				},
			},
			{
				ResourceName:                         "telnyx_box_tag.vip",
				ImportState:                          true,
				ImportStateId:                        "box-1/vip",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			{
				PreConfig: func() { api.set(func() { api.box["tags"] = []any{"other"} }) },
				Config:    cfg,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_box_tag.vip", plancheck.ResourceActionCreate)},
				},
			},
		},
	})
	if api.sent("DELETE /boxes/box-1/tags/vip") != 1 {
		t.Errorf("tag not deleted: %v", api.log())
	}
}

func TestExtraUpdates(t *testing.T) {
	api, c := newCollectionAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, box()),
		Steps: []rtest.TestStep{
			{
				Config: `resource "telnyx_box" "b" {
  name         = "b"
  blocklist_id = "bl-1"
}`,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckResourceAttr("telnyx_box.b", "blocklist_id", "bl-1"),
					func(*terraform.State) error {
						if api.sent(`POST /boxes? {"name":"b"}`) != 1 || api.sent(`POST /boxes/box-1/actions/set_blocklist? {"blocklist_id":"bl-1"}`) != 1 {
							return fmt.Errorf("unexpected create requests: %v", api.log())
						}
						return nil
					},
				),
			},
			{
				// Only the name changes: no action is called.
				Config: `resource "telnyx_box" "b" {
  name         = "c"
  blocklist_id = "bl-1"
}`,
				Check: func(*terraform.State) error {
					if api.sent(`PATCH /boxes/box-1? {"name":"c"}`) != 1 || api.sent("POST /boxes/box-1/actions/") != 1 {
						return fmt.Errorf("unexpected update requests: %v", api.log())
					}
					return nil
				},
			},
			{
				// Only the blocklist changes: only its action is called.
				Config: `resource "telnyx_box" "b" {
  name         = "c"
  blocklist_id = "bl-2"
}`,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_box.b", plancheck.ResourceActionUpdate)},
				},
				Check: func(*terraform.State) error {
					if api.sent("PATCH") != 1 || api.sent(`POST /boxes/box-1/actions/set_blocklist? {"blocklist_id":"bl-2"}`) != 1 {
						return fmt.Errorf("unexpected update requests: %v", api.log())
					}
					return nil
				},
			},
			{
				// Removed: the clear action is called.
				Config: `resource "telnyx_box" "b" {
  name = "c"
}`,
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckNoResourceAttr("telnyx_box.b", "blocklist_id"),
					func(*terraform.State) error {
						if api.sent("POST /boxes/box-1/actions/remove_blocklist") != 1 {
							return fmt.Errorf("clear not sent: %v", api.log())
						}
						return nil
					},
				),
			},
		},
	})
}

func TestCollectionSchemasAreValid(t *testing.T) {
	ctx := context.Background()
	for _, d := range []*Resource{secret(), boxTag(), box()} {
		if diags := d.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", d.TypeName, diags)
		}
	}
	for _, l := range []*List{thingsList(), regionsList()} {
		if diags := l.Schema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", l.TypeName, diags)
		}
	}
}

func TestDecodeList(t *testing.T) {
	items, total, err := decodeList([]byte(`[{"id":"a"},"b"]`), false)
	if err != nil || len(items) != 2 || items[0]["id"] != "a" || items[1]["value"] != "b" || total != 0 {
		t.Errorf("bare array: %v %d %v", items, total, err)
	}
	items, total, err = decodeList([]byte(`{"data":[{"id":"a"}],"meta":{"total_pages":3}}`), true)
	if err != nil || len(items) != 1 || total != 3 {
		t.Errorf("envelope: %v %d %v", items, total, err)
	}
	if _, _, err := decodeList([]byte(`{"data":{}}`), true); err == nil {
		t.Error("an object for data must be an error")
	}
}
