// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package codegen

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// maxDepth is the deepest object nesting rendered as nested attributes;
// deeper objects become JSON strings.
const maxDepth = 4

var addKinds = map[string]engine.Kind{
	"string": engine.KindString,
	"int":    engine.KindInt,
	"float":  engine.KindFloat,
	"bool":   engine.KindBool,
	"json":   engine.KindJSON,
}

// reserved are names Terraform does not allow as top-level attributes, plus
// the engine's timeouts block.
var reserved = map[string]bool{
	"count": true, "depends_on": true, "for_each": true, "lifecycle": true,
	"provider": true, "provisioner": true, "connection": true, "timeouts": true,
}

// Generator derives descriptors from a spec.
type Generator struct {
	Spec *apisync.Spec
}

type operation struct {
	method, path string
	node         map[string]any
}

func (g *Generator) operation(key string) (*operation, error) {
	if key == "" {
		return nil, nil
	}
	method, p, _ := strings.Cut(key, " ")
	item, ok := g.Spec.Paths()[p].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: path not in the spec", key)
	}
	n, ok := item[strings.ToLower(method)].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: operation not in the spec", key)
	}
	return &operation{method: method, path: p, node: n}, nil
}

// resolve follows $ref chains.
func (g *Generator) resolve(n any) map[string]any {
	for i := 0; i < 32; i++ {
		m, ok := n.(map[string]any)
		if !ok {
			return nil
		}
		ref, ok := m["$ref"].(string)
		if !ok {
			return m
		}
		target, ok := g.Spec.Component(ref)
		if !ok {
			return nil
		}
		n = target
	}
	return nil
}

// requestSchema returns the request body schema, preferring JSON.
func (g *Generator) requestSchema(op *operation) map[string]any {
	if op == nil {
		return nil
	}
	body := g.resolve(op.node["requestBody"])
	content, _ := body["content"].(map[string]any)
	for _, ct := range []string{"application/json", "multipart/form-data"} {
		if c, ok := content[ct].(map[string]any); ok {
			return g.norm(c["schema"])
		}
	}
	return nil
}

// responseSchema returns the first 2xx JSON response schema and whether the
// object is wrapped in "data".
func (g *Generator) responseSchema(op *operation) (map[string]any, bool) {
	if op == nil {
		return nil, false
	}
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
		r := g.resolve(responses[code])
		content, _ := r["content"].(map[string]any)
		c, ok := content["application/json"].(map[string]any)
		if !ok {
			continue
		}
		s := g.norm(c["schema"])
		props := properties(s)
		if len(props) == 1 && props["data"] != nil {
			return g.norm(props["data"]), true
		}
		return s, false
	}
	return nil, false
}

// norm resolves a schema and flattens allOf, nullable types and unions to a
// single shape. Unions without a stable shape are marked x-json.
func (g *Generator) norm(n any) map[string]any {
	s := g.resolve(n)
	if s == nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range s {
		out[k] = v
	}
	if parts, ok := s["allOf"].([]any); ok {
		delete(out, "allOf")
		props := map[string]any{}
		var req []any
		for _, p := range parts {
			ps := g.norm(p)
			for k, v := range properties(ps) {
				props[k] = v
			}
			req = append(req, asSlice(ps["required"])...)
			if out["description"] == nil && ps["description"] != nil {
				out["description"] = ps["description"]
			}
			if ps["x-json"] == true {
				out["x-json"] = true
			}
		}
		for k, v := range properties(s) {
			props[k] = v
		}
		req = append(req, asSlice(s["required"])...)
		out["type"] = "object"
		out["properties"] = props
		out["required"] = req
	}
	if ts, ok := s["type"].([]any); ok {
		out["type"] = nil
		for _, t := range ts {
			if t != "null" {
				out["type"] = t
				break
			}
		}
	}
	for _, key := range []string{"oneOf", "anyOf"} {
		vs, ok := s[key].([]any)
		if !ok {
			continue
		}
		delete(out, key)
		delete(out, "discriminator")
		var variants []map[string]any
		for _, v := range vs {
			nv := g.norm(v)
			if nv == nil || nv["type"] == "null" {
				continue
			}
			variants = append(variants, nv)
		}
		merged := g.mergeVariants(variants)
		if merged == nil {
			out["x-json"] = true
			continue
		}
		for k, v := range merged {
			if k == "description" && out["description"] != nil {
				continue
			}
			out[k] = v
		}
	}
	return out
}

// mergeVariants merges union variants that share one shape: a single
// variant, scalars of one type, or objects told apart by a single-valued
// "type" property (the Telnyx convention for tools and similar unions).
func (g *Generator) mergeVariants(vs []map[string]any) map[string]any {
	switch len(vs) {
	case 0:
		return map[string]any{"x-json": true}
	case 1:
		return vs[0]
	}
	first := vs[0]["type"]
	scalars := true
	for _, v := range vs {
		t, _ := v["type"].(string)
		if t == "" || t == "object" || t == "array" || v["type"] != first || v["x-json"] == true {
			scalars = false
		}
	}
	if scalars {
		out := map[string]any{"type": first}
		var enum []any
		for _, v := range vs {
			e, ok := v["enum"].([]any)
			if !ok {
				enum = nil
				break
			}
			enum = append(enum, e...)
		}
		if enum != nil {
			out["enum"] = enum
		}
		return out
	}
	props := map[string]any{}
	var tags []any
	for _, v := range vs {
		p := properties(v)
		tag := g.norm(p["type"])
		values := enumValues(tag)
		if len(values) != 1 {
			return nil
		}
		tags = append(tags, values[0])
		for k, pv := range p {
			if k != "type" {
				props[k] = pv
			}
		}
	}
	props["type"] = map[string]any{"type": "string", "enum": tags}
	return map[string]any{"type": "object", "properties": props, "required": []any{"type"}}
}

func enumValues(s map[string]any) []any {
	if c, ok := s["const"]; ok {
		return []any{c}
	}
	e, _ := s["enum"].([]any)
	return e
}

func properties(s map[string]any) map[string]any {
	p, _ := s["properties"].(map[string]any)
	return p
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

func required(s map[string]any, name string) bool {
	for _, r := range asSlice(s["required"]) {
		if r == name {
			return true
		}
	}
	return false
}

var paramPattern = regexp.MustCompile(`\{([^}]+)\}`)

func pathParams(p string) []string {
	var out []string
	for _, m := range paramPattern.FindAllStringSubmatch(p, -1) {
		out = append(out, m[1])
	}
	return out
}

// sources holds the schemas one attribute appears in.
type sources struct {
	create, update, read map[string]any
}

func (s sources) child(name string, g *Generator) sources {
	pick := func(parent map[string]any, request bool) map[string]any {
		c := g.norm(properties(parent)[name])
		if c == nil {
			return nil
		}
		if request && c["readOnly"] == true {
			return nil
		}
		if !request && c["writeOnly"] == true {
			return nil
		}
		return c
	}
	return sources{create: pick(s.create, true), update: pick(s.update, true), read: pick(s.read, false)}
}

func (s sources) items(g *Generator) sources {
	it := func(x map[string]any) map[string]any {
		if x == nil {
			return nil
		}
		return g.norm(x["items"])
	}
	return sources{create: it(s.create), update: it(s.update), read: it(s.read)}
}

func (s sources) shape() map[string]any {
	for _, x := range []map[string]any{s.create, s.update, s.read} {
		if x != nil {
			return x
		}
	}
	return nil
}

func (s sources) names() []string {
	set := map[string]bool{}
	for _, x := range []map[string]any{s.create, s.update, s.read} {
		for k := range properties(x) {
			set[k] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// deriver carries the state of one resource's derivation.
type deriver struct {
	g            *Generator
	ov           *Override
	used         map[string]bool
	hasUpdate    bool
	updateMethod string
}

// attrs derives the attributes of one object level.
func (d *deriver) attrs(src sources, depth int, parentComputed bool, prefix string) ([]*engine.Attr, error) {
	var out []*engine.Attr
	names := map[string]string{}
	for _, name := range src.names() {
		p := prefix + name
		aov := d.ov.Attributes[p]
		if _, ok := d.ov.Attributes[p]; ok {
			d.used[p] = true
		}
		if aov.Skip {
			continue
		}
		cs := src.child(name, d.g)
		if aov.Returned && cs.read == nil {
			cs.read = cs.shape()
		}
		inC, inU, inR := cs.create != nil, cs.update != nil, cs.read != nil
		configurable := inC || inU
		if !configurable && !inR {
			continue
		}
		shape := cs.shape()
		a := &engine.Attr{
			Name:        snake(name),
			JSON:        name,
			InCreate:    inC,
			InUpdate:    inU,
			InResponse:  inR,
			Sensitive:   aov.Sensitive,
			Description: describe(cs, aov),
		}
		if aov.Name != "" {
			a.Name = aov.Name
		}
		if shape["deprecated"] == true {
			a.Deprecated = "Deprecated by Telnyx."
		}
		a.Required = inC && required(src.create, name) && !parentComputed
		if aov.Required != nil {
			a.Required = *aov.Required
		}
		if !a.Required {
			a.Optional = configurable
			a.Computed = inR
		}
		// A PUT body often repeats the whole object, server fields included.
		// Fields only a PUT accepts are the server's: computed, and sent back
		// unchanged.
		if d.updateMethod == "PUT" && !inC && inU && inR {
			a.Optional = false
		}
		if aov.Computed != nil {
			a.Computed = *aov.Computed
			if a.Computed && !configurable {
				a.Optional = false
			}
		}
		if depth == 0 && configurable {
			a.ForceNew = !d.hasUpdate || !inU
		}
		if aov.ForceNew != nil {
			a.ForceNew = *aov.ForceNew
		}
		if err := d.kind(a, cs, depth, parentComputed || a.Computed, p, aov); err != nil {
			return nil, err
		}
		if prev, ok := names[a.Name]; ok {
			return nil, fmt.Errorf("%s and %s both map to attribute %s; rename one", prev, p, a.Name)
		}
		if depth == 0 && reserved[a.Name] {
			return nil, fmt.Errorf("%s maps to %s, which Terraform reserves; set a name in attributes.%s", p, a.Name, p)
		}
		names[a.Name] = p
		out = append(out, a)
	}
	return out, nil
}

// kind sets the attribute's Terraform shape.
func (d *deriver) kind(a *engine.Attr, cs sources, depth int, computed bool, p string, aov AttrOverride) error {
	shape := cs.shape()
	if aov.JSON || shape["x-json"] == true {
		a.Kind = engine.KindJSON
		return nil
	}
	t, _ := shape["type"].(string)
	if t == "" && properties(shape) != nil {
		t = "object"
	}
	switch t {
	case "string":
		a.Kind = engine.KindString
		if a.Optional || a.Required {
			a.Enum = stringEnum(cs.create, cs.update)
		}
	case "integer":
		a.Kind = engine.KindInt
	case "number":
		a.Kind = engine.KindFloat
	case "boolean":
		a.Kind = engine.KindBool
	case "object":
		if depth+1 >= maxDepth || len(properties(shape)) == 0 {
			a.Kind = engine.KindJSON
			return nil
		}
		children, err := d.attrs(cs, depth+1, computed, p+".")
		if err != nil {
			return err
		}
		if len(children) == 0 {
			a.Kind = engine.KindJSON
			return nil
		}
		a.Kind = engine.KindObject
		a.Attrs = children
	case "array":
		a.Kind = engine.KindList
		if aov.Set {
			a.Kind = engine.KindSet
		}
		items := cs.items(d.g)
		elem := &engine.Attr{Name: a.Name, InCreate: a.InCreate, InUpdate: a.InUpdate, InResponse: a.InResponse}
		if err := d.kind(elem, items, depth, computed, p+"[]", AttrOverride{}); err != nil {
			return err
		}
		switch elem.Kind {
		case engine.KindObject:
			a.Attrs = elem.Attrs
		case engine.KindList, engine.KindSet:
			a.Kind = engine.KindJSON // arrays of arrays
		default:
			a.Elem = &engine.Attr{Kind: elem.Kind}
		}
	default:
		a.Kind = engine.KindJSON
	}
	return nil
}

func stringEnum(schemas ...map[string]any) []string {
	set := map[string]bool{}
	for _, s := range schemas {
		if s == nil {
			continue
		}
		e, ok := s["enum"].([]any)
		if !ok {
			return nil // a request without an enum accepts any value
		}
		for _, v := range e {
			if str, ok := v.(string); ok {
				set[str] = true
			}
		}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var spaces = regexp.MustCompile(`\s+`)

func describe(cs sources, aov AttrOverride) string {
	if aov.Description != "" {
		return aov.Description
	}
	for _, s := range []map[string]any{cs.read, cs.create, cs.update} {
		if d, ok := s["description"].(string); ok && d != "" {
			return strings.TrimSpace(spaces.ReplaceAllString(d, " "))
		}
	}
	return ""
}

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return strings.Trim(nonWord.ReplaceAllString(b.String(), "_"), "_")
}

// Derive builds the descriptor for one override.
func (g *Generator) Derive(ov *Override) (*engine.Resource, error) {
	ops := map[string]*operation{}
	for name, key := range map[string]string{"create": ov.Create, "read": ov.Read, "update": ov.Update, "delete": ov.Delete} {
		op, err := g.operation(key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", ov.Type, name, err)
		}
		ops[name] = op
	}
	create, read, update, del := ops["create"], ops["read"], ops["update"], ops["delete"]

	readSchema, readEnv := g.responseSchema(read)
	if readSchema == nil {
		return nil, fmt.Errorf("%s: %s has no JSON response schema", ov.Type, ov.Read)
	}
	_, createEnv := g.responseSchema(create)
	src := sources{create: g.requestSchema(create), read: readSchema}
	if update != nil {
		src.update = g.requestSchema(update)
	}

	d := &deriver{g: g, ov: ov, used: map[string]bool{}, hasUpdate: update != nil}
	if update != nil {
		d.updateMethod = update.method
	}
	attrs, err := d.attrs(src, 0, false, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	r := &engine.Resource{
		TypeName:      ov.Type,
		Description:   ov.Description,
		Create:        &engine.Op{Method: create.method, Path: create.path, Envelope: createEnv},
		Read:          &engine.Op{Method: read.method, Path: read.path, Envelope: readEnv},
		Attrs:         attrs,
		ExactlyOne:    ov.ExactlyOne,
		DeleteWarning: ov.DeleteWarning,
		Timeouts:      ov.Timeouts,
		DataSource:    ov.dataSource(),
	}
	if update != nil {
		_, env := g.responseSchema(update)
		r.Update = &engine.Op{Method: update.method, Path: update.path, Envelope: env}
	}
	if del != nil {
		r.Delete = &engine.Op{Method: del.method, Path: del.path}
	}
	if err := g.ids(r, ov, create, []*operation{read, update, del}); err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	for _, add := range ov.Add {
		if r.Attr(add.Name) != nil {
			return nil, fmt.Errorf("%s: added attribute %s already exists", ov.Type, add.Name)
		}
		r.Attrs = append(r.Attrs, &engine.Attr{
			Name: add.Name, Kind: addKinds[add.Kind], Description: add.Description, Local: true,
			Required: add.Required, Optional: add.Optional, Computed: add.Computed,
			Sensitive: add.Sensitive, ForceNew: add.ForceNew,
		})
	}
	var unused []string
	for p := range ov.Attributes {
		if !d.used[p] {
			unused = append(unused, p)
		}
	}
	if len(unused) > 0 {
		sort.Strings(unused)
		return nil, fmt.Errorf("%s: attributes not in the spec: %s", ov.Type, strings.Join(unused, ", "))
	}
	for _, group := range ov.ExactlyOne {
		for _, n := range group {
			if r.Attr(n) == nil {
				return nil, fmt.Errorf("%s: exactly_one names unknown attribute %s", ov.Type, n)
			}
		}
	}
	sort.SliceStable(r.Attrs, func(i, j int) bool { return r.Attrs[i].Name < r.Attrs[j].Name })
	return r, nil
}

// ids wires the id attribute and parent path parameters, and sets Import.
func (g *Generator) ids(r *engine.Resource, ov *Override, create *operation, others []*operation) error {
	createParams := pathParams(create.path)
	rename := func(p string) string {
		if n, ok := ov.PathParams[p]; ok {
			return n
		}
		return snake(p)
	}
	// Parent parameters, and for settings resources the object's own id,
	// come from the create path.
	var parents []*engine.Attr
	for _, p := range createParams {
		name := rename(p)
		a := r.Attr(name)
		if a == nil {
			a = &engine.Attr{Name: name, Kind: engine.KindString, Local: true}
			r.Attrs = append(r.Attrs, a)
		}
		a.Required, a.Optional, a.Computed, a.ForceNew = true, false, false, true
		a.InCreate, a.InUpdate = false, false
		a.Kind = engine.KindString
		a.PathParams = append(a.PathParams, p)
		if a.Description == "" {
			a.Description = "The " + strings.ReplaceAll(strings.TrimSuffix(name, "_id"), "_", " ") + " id."
		}
		parents = append(parents, a)
	}

	var id *engine.Attr
	if ov.IDFromPath != "" {
		i := slices.Index(createParams, ov.IDFromPath)
		if i < 0 {
			return fmt.Errorf("id_from_path %s is not a parameter of %s", ov.IDFromPath, create.path)
		}
		id = parents[i]
		r.IDAttr = id.Name
		r.Import = []string{id.Name}
		for _, op := range others {
			if op == nil {
				continue
			}
			ps := pathParams(op.path)
			for j, p := range ps {
				if j < len(parents) && !slices.Contains(parents[j].PathParams, p) {
					parents[j].PathParams = append(parents[j].PathParams, p)
				}
			}
		}
		return nil
	}

	field := ov.ID
	if field == "" {
		field = "id"
	}
	for _, a := range r.Attrs {
		if a.JSON == field {
			id = a
		}
	}
	if id == nil {
		return fmt.Errorf("read response has no %s field; set id or id_from_path", field)
	}
	if other := r.Attr("id"); other != nil && other != id {
		return fmt.Errorf("response field %s is the id, but another attribute is named id; rename it", field)
	}
	id.Name = "id"
	id.Required, id.Optional, id.Computed, id.ForceNew = false, false, true, false
	id.InCreate, id.InUpdate, id.InResponse = false, false, true
	id.Kind = engine.KindString
	if id.Description == "" {
		id.Description = "The object's id."
	}
	r.IDAttr = "id"
	for _, op := range others {
		if op == nil {
			continue
		}
		ps := pathParams(op.path)
		if len(ps) == 0 {
			return fmt.Errorf("%s %s has no id parameter", op.method, op.path)
		}
		for j, p := range ps {
			target := id
			if j < len(ps)-1 {
				if j >= len(parents) {
					return fmt.Errorf("%s %s has a parameter %s that create does not", op.method, op.path, p)
				}
				target = parents[j]
			}
			if !slices.Contains(target.PathParams, p) {
				target.PathParams = append(target.PathParams, p)
			}
		}
	}
	for _, p := range parents {
		r.Import = append(r.Import, p.Name)
	}
	r.Import = append(r.Import, "id")
	return nil
}
