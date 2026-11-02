// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// conformanceSkip lists resources that the conformance test does not run,
// each with its reason. Their own tests in resources_test.go cover them.
var conformanceSkip = map[string]string{
	"telnyx_ai_avatar":    "creates by multipart upload",
	"telnyx_ai_tool":      "its settings must match its type",
	"telnyx_number_order": "waits for the order to complete",
}

// TestConformance runs every generated resource through create, an empty
// plan, its data source, import and destroy against a fake API driven by
// the resource's own descriptor. It checks the wiring the generator
// derives (paths, ids, envelopes, import ids), not the API's behaviour.
func TestConformance(t *testing.T) {
	for _, d := range descriptors() {
		if d.NoResource {
			continue
		}
		if why, ok := conformanceSkip[d.TypeName]; ok {
			t.Logf("%s: skipped, %s", d.TypeName, why)
			continue
		}
		t.Run(d.TypeName, func(t *testing.T) {
			t.Parallel()
			conform(t, d)
		})
	}
}

func conform(t *testing.T, d *engine.Resource) {
	api := newDescriptorAPI(t, d)
	name := d.TypeName + ".x"

	var b strings.Builder
	fmt.Fprintf(&b, "%s\nresource %q \"x\" {\n", api.config(), d.TypeName)
	for _, line := range requiredConfig(d) {
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("}\n")
	if d.DataSource {
		fmt.Fprintf(&b, "data %q \"x\" {\n", d.TypeName)
		for _, n := range d.Import {
			fmt.Fprintf(&b, "  %s = %s.%s\n", n, name, n)
		}
		b.WriteString("}\n")
	}

	// Attributes that import cannot recover: local values (other than
	// the import id parts) and values the API never returns.
	var ignore []string
	for _, a := range d.Attrs {
		if slices.Contains(d.Import, a.Name) {
			continue
		}
		if a.Local || a.WriteOnly || !a.InResponse {
			ignore = append(ignore, a.Name)
		}
	}
	if d.Timeouts {
		ignore = append(ignore, "timeouts")
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy: func(*terraform.State) error {
			api.mu.Lock()
			defer api.mu.Unlock()
			if d.Delete != nil && !d.ReleaseOnDestroy && api.obj != nil {
				return fmt.Errorf("destroy left the object in the API")
			}
			if (d.Delete == nil || d.ReleaseOnDestroy) && api.deleted {
				return fmt.Errorf("destroy deleted an object it should keep")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: b.String(),
				Check: func(s *terraform.State) error {
					if d.IDAttr == "" {
						return nil
					}
					return resource.TestCheckResourceAttrSet(name, d.IDAttr)(s)
				},
			},
			{
				ResourceName: name,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[name]
					parts := make([]string, 0, len(d.Import))
					for _, n := range d.Import {
						parts = append(parts, rs.Primary.Attributes[n])
					}
					return strings.Join(parts, "/"), nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: d.IDAttr,
				ImportStateVerifyIgnore:              ignore,
			},
		},
	})
}

// requiredConfig returns "name = value" lines for every required
// top-level attribute, and the first attribute of each exactly-one group.
func requiredConfig(d *engine.Resource) []string {
	var out []string
	for _, a := range d.Attrs {
		first := false
		for _, g := range d.ExactlyOne {
			first = first || len(g) > 0 && g[0] == a.Name
		}
		if a.Required || first {
			out = append(out, a.Name+" = "+sampleValue(a))
		}
	}
	return out
}

// sampleValue is a valid HCL value for an attribute: the first enum value,
// a placeholder of the right kind, or an object of its required fields.
func sampleValue(a *engine.Attr) string {
	switch a.Kind {
	case engine.KindString:
		if len(a.Enum) > 0 {
			return fmt.Sprintf("%q", a.Enum[0])
		}
		return fmt.Sprintf("%q", "v-"+a.Name)
	case engine.KindJSON:
		return `jsonencode({ key = "value" })`
	case engine.KindInt:
		return "1"
	case engine.KindFloat:
		return "1.5"
	case engine.KindBool:
		return "true"
	case engine.KindList, engine.KindSet:
		if a.Elem != nil {
			return "[" + sampleValue(&engine.Attr{Name: a.Name, Kind: a.Elem.Kind}) + "]"
		}
		return "[" + sampleObject(a.Attrs) + "]"
	case engine.KindObject:
		return sampleObject(a.Attrs)
	}
	return "null"
}

func sampleObject(attrs []*engine.Attr) string {
	var parts []string
	for _, c := range attrs {
		if c.Required {
			parts = append(parts, c.Name+" = "+sampleValue(c))
		}
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// descriptorAPI is a fake API for one resource. It holds at most one
// object and routes requests by the descriptor's operations.
type descriptorAPI struct {
	t   *testing.T
	d   *engine.Resource
	url string

	mu      sync.Mutex
	obj     map[string]any
	deleted bool
}

type route struct {
	op   *engine.Op
	kind string // create, read, update, delete or list
	re   *regexp.Regexp
	keys []string
}

func newDescriptorAPI(t *testing.T, d *engine.Resource) *descriptorAPI {
	t.Helper()
	a := &descriptorAPI{t: t, d: d}
	srv := httptest.NewServer(a)
	t.Cleanup(srv.Close)
	a.url = srv.URL
	return a
}

func (a *descriptorAPI) config() string {
	return fmt.Sprintf(`
provider "telnyx" {
  api_key     = "test"
  base_url    = %q
  max_retries = 0
}
`, a.url)
}

func (a *descriptorAPI) factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return (&fakeTelnyx{}).factories()
}

// newRoute matches an operation's path, capturing its parameters.
func newRoute(op *engine.Op, kind string) route {
	var keys []string
	segs := strings.Split(op.Path, "/")
	for i, seg := range segs {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			keys = append(keys, seg[1:len(seg)-1])
			segs[i] = `([^/]+)`
		} else {
			segs[i] = regexp.QuoteMeta(seg)
		}
	}
	return route{op: op, kind: kind, re: regexp.MustCompile("^" + strings.Join(segs, "/") + "$"), keys: keys}
}

func (a *descriptorAPI) routes() []route {
	d := a.d
	var rs []route
	add := func(op *engine.Op, kind string) {
		if op != nil {
			rs = append(rs, newRoute(op, kind))
		}
	}
	add(d.Create, "create")
	add(d.Update, "update")
	add(d.Read, "read")
	add(d.Delete, "delete")
	if d.ReadFrom != nil {
		add(d.ReadFrom.Op, "list")
	}
	for _, u := range d.Updates {
		add(u.Op, "update")
		add(u.Clear, "update")
	}
	return rs
}

func (a *descriptorAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)

	for _, rt := range a.routes() {
		if rt.op.Method != r.Method {
			continue
		}
		m := rt.re.FindStringSubmatch(r.URL.Path)
		if m == nil {
			continue
		}
		params := map[string]string{}
		for i, k := range rt.keys {
			params[k] = m[i+1]
		}
		kind := rt.kind
		// Settings objects create with their update operation.
		if (kind == "create" || kind == "update") && a.obj == nil {
			kind = "create"
		} else if kind == "create" && rt.op == a.d.Update {
			kind = "update"
		}
		switch kind {
		case "create":
			a.create(w, rt, body, params)
		case "update":
			if a.obj == nil {
				http.NotFound(w, r)
				return
			}
			if rt.op.Method == http.MethodPut {
				keep := a.obj
				a.obj = body
				for k, v := range keep {
					if _, ok := a.obj[k]; !ok {
						a.obj[k] = v
					}
				}
			} else {
				for k, v := range body {
					a.obj[k] = v
				}
			}
			a.reply(w, rt.op, a.obj)
		case "read":
			if a.obj == nil {
				http.NotFound(w, r)
				return
			}
			a.reply(w, rt.op, a.obj)
		case "list":
			items := []any{}
			if a.obj != nil {
				items = append(items, a.obj)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"data": items, "meta": map[string]any{"total_pages": 1}})
		case "delete":
			a.obj, a.deleted = nil, true
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}
	a.t.Errorf("%s: unexpected request %s %s", a.d.TypeName, r.Method, r.URL.Path)
	http.Error(w, `{"errors":[{"detail":"unexpected"}]}`, http.StatusTeapot)
}

func (a *descriptorAPI) create(w http.ResponseWriter, rt route, body map[string]any, params map[string]string) {
	d := a.d
	asList := false
	// Batch creates (a list of one) store and return that one element.
	if docs, ok := body["additional_documents"].([]any); ok && len(docs) == 1 {
		body, _ = docs[0].(map[string]any)
		asList = true
	}
	obj := map[string]any{}
	for k, v := range body {
		obj[k] = v
	}
	// The object's id: the path parameter of a settings object, or a new
	// value for objects that have their own id.
	id := "obj-1"
	if ida := d.Attr(d.IDAttr); ida != nil {
		for _, p := range ida.PathParams {
			if v, ok := params[p]; ok {
				id = v
			}
		}
	}
	if ida := d.Attr(d.IDAttr); ida != nil && ida.JSON != "" && !ida.Local {
		if _, ok := obj[ida.JSON]; !ok {
			obj[ida.JSON] = id
		}
	}
	if d.ReadFrom != nil && d.ReadFrom.Key != "" {
		if _, ok := obj[d.ReadFrom.Key]; !ok {
			obj[d.ReadFrom.Key] = id
		}
	}
	// Response fields that echo a path parameter, such as a parent id.
	for _, at := range d.Attrs {
		if at.InResponse && at.JSON != "" && len(at.PathParams) > 0 {
			for _, p := range at.PathParams {
				if v, ok := params[p]; ok {
					if _, set := obj[at.JSON]; !set {
						obj[at.JSON] = v
					}
				}
			}
		}
	}
	a.obj = obj
	if asList {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{obj}})
		return
	}
	a.reply(w, rt.op, obj)
}

func (a *descriptorAPI) reply(w http.ResponseWriter, op *engine.Op, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	if op.Envelope {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": obj})
		return
	}
	_ = json.NewEncoder(w).Encode(obj)
}

// TestConformanceSkipList keeps the skip list honest: every skipped
// resource exists.
func TestConformanceSkipList(t *testing.T) {
	names := map[string]bool{}
	for _, d := range descriptors() {
		names[d.TypeName] = true
	}
	keys := make([]string, 0, len(conformanceSkip))
	for k := range conformanceSkip {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !names[k] {
			t.Errorf("conformanceSkip names %s, which is not a resource", k)
		}
	}
}
