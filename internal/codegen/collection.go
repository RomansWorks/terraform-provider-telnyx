// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// DeriveAll derives what one override describes: a resource (or a
// data-source-only descriptor) and, when the override names a list
// operation, a list data source. Either may be nil.
func (g *Generator) DeriveAll(ov *Override) (*engine.Resource, *engine.List, error) {
	var r *engine.Resource
	var err error
	switch {
	case ov.Kind == "":
		r, err = g.Derive(ov)
	case ov.Read != "":
		r, err = g.deriveData(ov)
	}
	if err != nil {
		return nil, nil, err
	}
	if ov.List == "" {
		return r, nil, nil
	}
	desc := ov.Description
	if ov.Kind == "" || ov.Read != "" {
		desc = "Lists " + strings.ReplaceAll(strings.TrimPrefix(Plural(ov.Type), "telnyx_"), "_", " ") +
			", with optional filters. Each item has the attributes of `" + ov.Type + "`."
	}
	l, err := g.DeriveList(ov.listName(), desc, ov.List, ov.Paging)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: list: %w", ov.Type, err)
	}
	return r, l, nil
}

// deriveData derives a data source with no resource from a read operation.
// Its path parameters are the required lookup attributes.
func (g *Generator) deriveData(ov *Override) (*engine.Resource, error) {
	read, err := g.operation(ov.Read)
	if err != nil {
		return nil, fmt.Errorf("%s: read: %w", ov.Type, err)
	}
	schema, env := g.responseSchema(read)
	if schema == nil {
		return nil, fmt.Errorf("%s: %s has no JSON response schema", ov.Type, ov.Read)
	}
	d := &deriver{g: g, ov: ov, used: map[string]bool{}}
	attrs, err := d.attrs(sources{read: schema}, 0, false, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	r := &engine.Resource{
		TypeName:    ov.Type,
		Description: ov.Description,
		Read:        &engine.Op{Method: read.method, Path: read.path, Envelope: env},
		Attrs:       attrs,
		DataSource:  true,
		NoResource:  true,
	}
	for _, p := range pathParams(read.path) {
		name := snake(p)
		if n, ok := ov.PathParams[p]; ok {
			name = n
		}
		a := r.Attr(name)
		if a == nil {
			a = &engine.Attr{Name: name, Kind: engine.KindString, Local: true,
				Description: "The " + strings.ReplaceAll(strings.TrimSuffix(name, "_id"), "_", " ") + "."}
			r.Attrs = append(r.Attrs, a)
		}
		a.Kind = engine.KindString
		a.PathParams = append(a.PathParams, p)
		r.Import = append(r.Import, name)
	}
	if err := checkUnused(ov, d); err != nil {
		return nil, err
	}
	sort.SliceStable(r.Attrs, func(i, j int) bool { return r.Attrs[i].Name < r.Attrs[j].Name })
	return r, nil
}

func checkUnused(ov *Override, d *deriver) error {
	var unused []string
	for p := range ov.Attributes {
		if !d.used[p] {
			unused = append(unused, p)
		}
	}
	if len(unused) > 0 {
		sort.Strings(unused)
		return fmt.Errorf("%s: attributes not in the spec: %s", ov.Type, strings.Join(unused, ", "))
	}
	return nil
}

// listItems returns the schema of one element of a list response's data
// array, and whether the array is wrapped in data (false for a bare array).
func (g *Generator) listItems(op *operation) (map[string]any, bool, error) {
	responses, _ := op.node["responses"].(map[string]any)
	codes := make([]string, 0, len(responses))
	for c := range responses {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	for _, code := range codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		content, _ := g.resolve(responses[code])["content"].(map[string]any)
		c, ok := content["application/json"].(map[string]any)
		if !ok {
			continue
		}
		schema := g.norm(c["schema"])
		if schema["type"] == "array" {
			return g.norm(schema["items"]), false, nil
		}
		data := g.norm(properties(schema)["data"])
		if data == nil || data["type"] != "array" {
			return nil, false, fmt.Errorf("%s %s does not return a data array", op.method, op.path)
		}
		return g.norm(data["items"]), true, nil
	}
	return nil, false, fmt.Errorf("%s %s has no JSON response", op.method, op.path)
}

// param is one operation parameter.
type param struct {
	name, in string
	schema   map[string]any
}

func (g *Generator) params(op *operation) []param {
	var raw []any
	if item, ok := g.Spec.Paths()[op.path].(map[string]any); ok {
		raw = append(raw, asSlice(item["parameters"])...)
	}
	raw = append(raw, asSlice(op.node["parameters"])...)
	var out []param
	for _, r := range raw {
		p := g.resolve(r)
		name, _ := p["name"].(string)
		in, _ := p["in"].(string)
		if name == "" {
			continue
		}
		out = append(out, param{name: name, in: in, schema: g.norm(p["schema"])})
	}
	return out
}

// pagingParams are the query parameters each paging style uses.
var pagingParams = map[string]engine.Paging{
	"page[number]": engine.PagingNumber, "page[size]": engine.PagingNumber,
	"page_number": engine.PagingFlat, "page_size": engine.PagingFlat,
}

// otherPaging are paging parameters the engine does not drive.
var otherPaging = []string{"page[after]", "page[before]", "page[limit]", "page[snapshot]", "page_cursor",
	"cursor", "limit", "offset", "recordsPerPage", "page"}

// DeriveList derives a list data source from a list operation.
func (g *Generator) DeriveList(name, description, key, pagingOverride string) (*engine.List, error) {
	op, err := g.operation(key)
	if err != nil {
		return nil, err
	}
	item, env, err := g.listItems(op)
	if err != nil {
		return nil, err
	}
	listOp, err := g.listOp(op, pagingOverride)
	if err != nil {
		return nil, err
	}
	listOp.Envelope = env
	l := &engine.List{TypeName: name, Description: description, Op: listOp}

	seen := map[string]string{}
	addParam := func(a *engine.Attr, from string) error {
		if prev, ok := seen[a.Name]; ok {
			return fmt.Errorf("query parameters %s and %s both map to %s", prev, from, a.Name)
		}
		if reserved[a.Name] || a.Name == "max_items" {
			return fmt.Errorf("query parameter %s maps to reserved name %s", from, a.Name)
		}
		seen[a.Name] = from
		l.Params = append(l.Params, a)
		return nil
	}
	for _, p := range g.params(op) {
		switch {
		case p.in == "path":
			a := &engine.Attr{Name: snake(p.name), Kind: engine.KindString, Required: true, PathParams: []string{p.name},
				Description: "The " + strings.ReplaceAll(strings.TrimSuffix(snake(p.name), "_id"), "_", " ") + " id."}
			if err := addParam(a, p.name); err != nil {
				return nil, err
			}
			continue
		case p.in != "query", isPaging(p):
			continue
		}
		attrs, err := g.queryAttrs(p.name, p.name, p.schema, 0)
		if err != nil {
			return nil, err
		}
		for _, a := range attrs {
			if err := addParam(a, a.JSON); err != nil {
				return nil, err
			}
		}
	}
	sort.SliceStable(l.Params, func(i, j int) bool { return l.Params[i].Name < l.Params[j].Name })

	items := &engine.Attr{Name: strings.TrimPrefix(name, "telnyx_"), JSON: "data", Kind: engine.KindList,
		Computed: true, InResponse: true, Description: "The objects, in the order the API returns them."}
	ov := &Override{Type: name}
	d := &deriver{g: g, ov: ov, used: map[string]bool{}}
	shape := item
	t, _ := shape["type"].(string)
	if t == "object" || properties(shape) != nil {
		children, err := d.attrs(sources{read: item}, 1, true, "")
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			items.Elem = &engine.Attr{Kind: engine.KindJSON}
		}
		items.Attrs = children
	} else {
		elem := &engine.Attr{Name: items.Name, InResponse: true}
		if err := d.kind(elem, sources{read: item}, 1, true, "[]", AttrOverride{}); err != nil {
			return nil, err
		}
		if elem.Kind == engine.KindObject || elem.Kind == engine.KindList || elem.Kind == engine.KindSet {
			elem.Kind = engine.KindJSON
		}
		items.Elem = &engine.Attr{Kind: elem.Kind}
	}
	l.Items = items
	return l, nil
}

// isPaging reports whether a query parameter pages: the engine sends it,
// or it is a style the engine does not drive.
func isPaging(p param) bool {
	_, known := pagingParams[p.name]
	return known || p.name == "page" && properties(p.schema) != nil || contains(otherPaging, p.name)
}

// listOp returns the engine operation of a list, with its paging style and
// page size. A paging parameter the engine does not drive is an error unless
// pagingOverride is "none".
func (g *Generator) listOp(op *operation, pagingOverride string) (*engine.Op, error) {
	paging := engine.PagingNone
	size := 0
	unsupported := func(name string) error {
		if pagingOverride == "none" {
			return nil
		}
		return fmt.Errorf("paging parameter %s is not supported; set paging: none and page in a hook", name)
	}
	for _, p := range g.params(op) {
		if p.in != "query" {
			continue
		}
		// A deepObject "page" with number and size is page[number]/page[size].
		if p.name == "page" && properties(p.schema) != nil {
			keys := make([]string, 0, len(properties(p.schema)))
			for k := range properties(p.schema) {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				full := "page[" + k + "]"
				st, ok := pagingParams[full]
				if !ok {
					if err := unsupported(full); err != nil {
						return nil, err
					}
					continue
				}
				paging = st
				if k == "size" {
					size = maxOf(g.norm(properties(p.schema)[k]))
				}
			}
			continue
		}
		if st, ok := pagingParams[p.name]; ok {
			paging = st
			if strings.Contains(p.name, "size") {
				size = maxOf(p.schema)
			}
			continue
		}
		if contains(otherPaging, p.name) {
			if err := unsupported(p.name); err != nil {
				return nil, err
			}
		}
	}
	o := &engine.Op{Method: op.method, Path: op.path, Envelope: true}
	if pagingOverride == "none" {
		paging = engine.PagingNone
	}
	o.Paging = paging
	if paging != engine.PagingNone {
		o.PageSize = engine.DefaultPageSize
		if size > 0 && size < o.PageSize {
			o.PageSize = size
		}
	}
	return o, nil
}

// queryAttrs turns one query parameter into optional attributes. A
// deepObject parameter such as filter becomes one attribute per leaf:
// filter[status] is filter_status and filter[name][contains] is
// filter_name_contains; an "eq" leaf adds nothing to the name.
func (g *Generator) queryAttrs(name, key string, s map[string]any, depth int) ([]*engine.Attr, error) {
	if props := properties(s); props != nil && depth < 3 {
		keys := make([]string, 0, len(props))
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var out []*engine.Attr
		for _, k := range keys {
			child := name + "_" + k
			if k == "eq" {
				child = name
			}
			as, err := g.queryAttrs(child, key+"["+k+"]", g.norm(props[k]), depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, as...)
		}
		return out, nil
	}
	a := &engine.Attr{Name: snake(name), JSON: key, Optional: true, Description: describe(sources{read: s}, AttrOverride{})}
	t, _ := s["type"].(string)
	switch t {
	case "integer":
		a.Kind = engine.KindInt
	case "number":
		a.Kind = engine.KindFloat
	case "boolean":
		a.Kind = engine.KindBool
	case "array":
		a.Kind = engine.KindList
		a.Elem = &engine.Attr{Kind: engine.KindString}
	default:
		a.Kind = engine.KindString
		a.Enum = stringEnum(s)
	}
	return []*engine.Attr{a}, nil
}

func maxOf(s map[string]any) int {
	switch v := s["maximum"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
