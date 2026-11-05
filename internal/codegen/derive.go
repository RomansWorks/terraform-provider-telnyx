// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"maps"
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
	// mergeUnions is the merge_variants setting of the override being
	// derived.
	mergeUnions bool
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
	for _, ct := range []string{"application/json", "multipart/form-data", formType} {
		if c, ok := content[ct].(map[string]any); ok {
			return g.norm(c["schema"])
		}
	}
	return nil
}

const formType = "application/x-www-form-urlencoded"

// formBody reports whether an operation takes only a URL-encoded form.
func (g *Generator) formBody(op *operation) bool {
	if op == nil {
		return false
	}
	content, _ := g.resolve(op.node["requestBody"])["content"].(map[string]any)
	_, form := content[formType]
	_, js := content["application/json"]
	return form && !js
}

const rawType = "application/octet-stream"

// textSchema is the object a plain-text or raw response is read as.
var textSchema = map[string]any{"type": "object", "properties": map[string]any{
	"value": map[string]any{"type": "string", "description": "The value the API returned."}}}

// rawBody reports whether an operation takes only raw bytes, which a
// resource's Send hook supplies.
func (g *Generator) rawBody(op *operation) bool {
	if op == nil {
		return false
	}
	content, _ := g.resolve(op.node["requestBody"])["content"].(map[string]any)
	_, raw := content[rawType]
	return raw && len(content) == 1
}

// checkBody fails when an operation has a request body the provider
// cannot send.
func (g *Generator) checkBody(op *operation) error {
	if op == nil {
		return nil
	}
	content, _ := g.resolve(op.node["requestBody"])["content"].(map[string]any)
	if len(content) > 0 && g.requestSchema(op) == nil && !g.rawBody(op) {
		types := make([]string, 0, len(content))
		for ct := range content {
			types = append(types, ct)
		}
		sort.Strings(types)
		return fmt.Errorf("%s %s: request body types %s are not supported", op.method, op.path, strings.Join(types, ", "))
	}
	return nil
}

// responseSchema returns the first 2xx JSON response schema and whether the
// object is wrapped in "data".
func (g *Generator) responseSchema(op *operation) (map[string]any, bool) {
	return g.responseSchemaWhere(op, func(string) bool { return true })
}

// successCodes returns the operation's 2xx response codes, lowest first.
func successCodes(op *operation) []string {
	responses, _ := op.node["responses"].(map[string]any)
	var codes []string
	for c := range responses {
		if strings.HasPrefix(c, "2") {
			codes = append(codes, c)
		}
	}
	sort.Strings(codes)
	return codes
}

// responseSchemaWhere is responseSchema over the 2xx codes keep accepts.
func (g *Generator) responseSchemaWhere(op *operation, keep func(code string) bool) (map[string]any, bool) {
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
		if !strings.HasPrefix(code, "2") || !keep(code) {
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
		if d := g.norm(props["data"]); d != nil && (d["type"] == "object" || properties(d) != nil) {
			// Telnyx wraps objects in "data"; siblings such as meta,
			// links or a repeated id are envelope, not the object.
			return d, true
		}
		if d := g.norm(props["data"]); d != nil && d["type"] == "array" {
			// A list of objects: the caller decides whether one element
			// is the object (a create that takes a batch) or it is an error.
			return g.norm(d["items"]), true
		}
		return s, false
	}
	return nil, false
}

// textResponse reports whether the first 2xx response is a plain-text string.
func (g *Generator) textResponse(op *operation) bool {
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
		c, ok := content["text/plain"].(map[string]any)
		if !ok {
			c, ok = content[rawType].(map[string]any)
		}
		return ok && g.norm(c["schema"])["type"] == "string"
	}
	return false
}

// dataList reports whether an operation's 2xx response wraps a list in
// "data" rather than one object.
func (g *Generator) dataList(op *operation) bool {
	if op == nil {
		return false
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
		content, _ := g.resolve(responses[code])["content"].(map[string]any)
		c, ok := content["application/json"].(map[string]any)
		if !ok {
			continue
		}
		d := g.norm(properties(g.norm(c["schema"]))["data"])
		return d != nil && d["type"] == "array"
	}
	return false
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
	// An enum of strings without a type is a string.
	if e, ok := out["enum"].([]any); ok && out["type"] == nil && len(e) > 0 {
		str := true
		for _, v := range e {
			_, isStr := v.(string)
			str = str && isStr
		}
		if str {
			out["type"] = "string"
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
			if m := g.mergeByFlag(vs); m != nil || !g.mergeUnions {
				return m
			}
			return g.mergeSameShape(vs)
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

// mergeByFlag merges object variants told apart by a boolean property with
// one value per variant, such as "unlimited", false in one variant and true
// in the other. The flag keeps its type without the enum;
// a field is required when every variant requires it. Nil when the variants
// have no such property.
func (g *Generator) mergeByFlag(vs []map[string]any) map[string]any {
	var flag string
	for name := range properties(vs[0]) {
		seen := map[any]bool{}
		ok := true
		for _, v := range vs {
			values := enumValues(g.norm(properties(v)[name]))
			if len(values) != 1 || seen[values[0]] {
				ok = false
				break
			}
			if _, isBool := values[0].(bool); !isBool {
				ok = false
				break
			}
			seen[values[0]] = true
		}
		if ok && (flag == "" || name < flag) {
			flag = name
		}
	}
	if flag == "" {
		return nil
	}
	props := map[string]any{}
	count := map[string]int{}
	for _, v := range vs {
		for k, pv := range properties(v) {
			if _, ok := props[k]; !ok {
				props[k] = pv
			}
		}
		for _, r := range asSlice(v["required"]) {
			if n, ok := r.(string); ok {
				count[n]++
			}
		}
	}
	f := g.norm(props[flag])
	delete(f, "enum")
	props[flag] = f
	var req []any
	for _, k := range slices.Sorted(maps.Keys(count)) {
		if count[k] == len(vs) {
			req = append(req, k)
		}
	}
	return map[string]any{"type": "object", "properties": props, "required": req}
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
		if request && (c["readOnly"] == true || readOnlyObject(g, c)) {
			return nil
		}
		if !request && c["writeOnly"] == true {
			return nil
		}
		return c
	}
	return sources{create: pick(s.create, true), update: pick(s.update, true), read: pick(s.read, false)}
}

// readOnlyObject reports whether every field of an object is read-only, so
// that a request cannot set any of it (a SIM card's status).
func readOnlyObject(g *Generator, s map[string]any) bool {
	props := properties(s)
	if len(props) == 0 {
		return false
	}
	for _, v := range props {
		if g.norm(v)["readOnly"] != true {
			return false
		}
	}
	return true
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
	// writeOnly collects the write-only attributes, which need a version.
	writeOnly []*engine.Attr
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
			Sensitive:   aov.Sensitive || aov.Masked || aov.CreateOnly,
			Masked:      aov.Masked,
			CreateOnly:  aov.CreateOnly,
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
			// A hook returns exactly what the user set, so the user's value
			// is the whole truth: not computed.
			a.Computed = inR && !aov.Returned
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
		if aov.WriteOnly {
			if err := d.writeOnlyAttr(a, p, depth, configurable); err != nil {
				return nil, err
			}
		}
		if aov.CreateOnly {
			if depth > 0 || configurable {
				return nil, fmt.Errorf("%s: create_only applies to top-level fields that requests do not take", p)
			}
			a.Required, a.Optional, a.Computed, a.ForceNew = false, false, true, false
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

// writeOnlyAttr turns a secret request field into <name>_wo: sent from
// configuration, never stored or read back.
func (d *deriver) writeOnlyAttr(a *engine.Attr, p string, depth int, configurable bool) error {
	switch {
	case depth > 0:
		return fmt.Errorf("%s: write_only applies to top-level fields only", p)
	case !configurable:
		return fmt.Errorf("%s: write_only needs a field the request takes", p)
	case a.Kind != engine.KindString && a.Kind != engine.KindJSON && a.Kind != engine.KindInt && a.Kind != engine.KindFloat && a.Kind != engine.KindBool:
		return fmt.Errorf("%s: write_only applies to scalar fields only", p)
	}
	base := a.Name
	a.Name = base + "_wo"
	a.Version = base + "_wo_version"
	a.WriteOnly, a.Sensitive = true, true
	a.Computed, a.InResponse, a.ForceNew = false, false, false
	a.Optional = !a.Required
	d.writeOnly = append(d.writeOnly, a)
	return nil
}

// versions returns the <name>_wo_version attribute of each write-only
// attribute. A new version replaces the object when no update takes the
// field.
func (d *deriver) versions() []*engine.Attr {
	var out []*engine.Attr
	for _, a := range d.writeOnly {
		out = append(out, &engine.Attr{Name: a.Version, Kind: engine.KindInt, Optional: true, Local: true,
			ForceNew:    !d.hasUpdate || !a.InUpdate,
			Description: "Change this number to send `" + a.Name + "` again; Terraform cannot detect a change to a value it does not store."})
	}
	return out
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
	g.mergeUnions = ov.MergeVariants
	defer func() { g.mergeUnions = false }()
	ops := map[string]*operation{}
	for name, key := range map[string]string{"create": ov.Create, "read": ov.Read, "update": ov.Update, "delete": ov.Delete} {
		op, err := g.operation(key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", ov.Type, name, err)
		}
		ops[name] = op
	}
	create, read, update, del := ops["create"], ops["read"], ops["update"], ops["delete"]

	var readSchema map[string]any
	var readEnv, readText bool
	var rf *engine.ReadFrom
	var collection *operation
	if ov.ReadFrom != nil {
		var err error
		if collection, err = g.operation(ov.ReadFrom.Op); err != nil {
			return nil, fmt.Errorf("%s: read_from: %w", ov.Type, err)
		}
		if readSchema, rf, err = g.collection(ov, collection); err != nil {
			return nil, fmt.Errorf("%s: read_from: %w", ov.Type, err)
		}
	} else {
		readSchema, readEnv = g.responseSchema(read)
		if readSchema == nil && g.textResponse(read) {
			// A read that answers with the raw value, such as a KV key,
			// is one string field.
			readSchema, readText = textSchema, true
		}
		if readSchema == nil {
			return nil, fmt.Errorf("%s: %s has no JSON response schema", ov.Type, ov.Read)
		}
		if g.dataList(read) {
			return nil, fmt.Errorf("%s: read %s returns a list; use list or read_from", ov.Type, ov.Read)
		}
	}
	readSchema, err := g.withCreateOnly(ov, readSchema, create)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	for _, op := range []*operation{create, update} {
		if err := g.checkBody(op); err != nil {
			return nil, fmt.Errorf("%s: %w", ov.Type, err)
		}
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
		Create:        &engine.Op{Method: create.method, Path: create.path, Envelope: createEnv, First: g.dataList(create), Form: g.formBody(create), Raw: g.rawBody(create)},
		ReadFrom:      rf,
		Attrs:         attrs,
		ExactlyOne:    ov.ExactlyOne,
		DeleteWarning: ov.DeleteWarning,
		Timeouts:      ov.Timeouts,
		DataSource:    ov.dataSource() && read != nil,
	}
	if read != nil {
		r.Read = &engine.Op{Method: read.method, Path: read.path, Envelope: readEnv, Text: readText}
	}
	if update != nil {
		_, env := g.responseSchema(update)
		r.Update = &engine.Op{Method: update.method, Path: update.path, Envelope: env, Form: g.formBody(update), Raw: g.rawBody(update)}
	}
	if del != nil {
		r.Delete = &engine.Op{Method: del.method, Path: del.path}
	}
	others := []*operation{read, update, del}
	r.Attrs = append(r.Attrs, d.versions()...)
	if ov.ReleaseOnDestroy {
		if r.Attr(engine.ReleaseAttr) != nil {
			return nil, fmt.Errorf("%s: the API already has a field named %s; rename it", ov.Type, engine.ReleaseAttr)
		}
		r.ReleaseOnDestroy = true
		r.Attrs = append(r.Attrs, &engine.Attr{Name: engine.ReleaseAttr, Kind: engine.KindBool, Optional: true, Local: true,
			Description: "Whether destroy deletes the object in Telnyx. By default destroy only removes it from Terraform state. " +
				"Destroy uses the value in state, so set it to true and apply before destroying."})
	}
	extra, err := g.extraUpdates(r, ov, d, readSchema)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	others = append(others, extra...)
	if err := g.ids(r, ov, create, others, collection); err != nil {
		return nil, fmt.Errorf("%s: %w", ov.Type, err)
	}
	if err := g.queryParams(r, ov); err != nil {
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
	if err := checkUnused(ov, d); err != nil {
		return nil, err
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

// withCreateOnly adds the create_only fields that only the create response
// lists to a copy of the read schema, so that they become attributes.
func (g *Generator) withCreateOnly(ov *Override, read map[string]any, create *operation) (map[string]any, error) {
	var names []string
	for p, a := range ov.Attributes {
		if !a.CreateOnly {
			continue
		}
		if strings.Contains(p, ".") {
			return nil, fmt.Errorf("create_only %s: create_only applies to top-level fields only", p)
		}
		if properties(read)[p] == nil {
			names = append(names, p)
		}
	}
	if len(names) == 0 {
		return read, nil
	}
	if create == nil {
		return nil, fmt.Errorf("create_only needs create")
	}
	sort.Strings(names)
	props := map[string]any{}
	for k, v := range properties(read) {
		props[k] = v
	}
	for _, n := range names {
		// A create can answer 200 for a replay without the value and 201
		// with it, so look in every 2xx response, lowest code first.
		var f any
		for _, code := range successCodes(create) {
			resp, _ := g.responseSchemaWhere(create, func(c string) bool { return c == code })
			if f = properties(resp)[n]; f != nil {
				break
			}
		}
		if f == nil {
			return nil, fmt.Errorf("create_only %s is in neither the read nor the create response", n)
		}
		props[n] = f
	}
	out := map[string]any{}
	for k, v := range read {
		out[k] = v
	}
	out["properties"] = props
	return out, nil
}

// collection returns the schema of one element of the collection an object
// is read from, shaped like a read response, and the engine's ReadFrom.
// Plain-value elements become an object with the id field alone.
func (g *Generator) collection(ov *Override, op *operation) (map[string]any, *engine.ReadFrom, error) {
	rf := &engine.ReadFrom{Field: ov.ReadFrom.Field, Key: ov.ReadFrom.Key}
	var item map[string]any
	if rf.Field == "" {
		o, err := g.listOp(op, "")
		if err != nil {
			return nil, nil, err
		}
		rf.Op = o
		var field string
		if item, o.Envelope, field, err = g.listItems(op); err != nil {
			return nil, nil, err
		}
		o.Records, o.Results = field == "records", field == "results"
	} else {
		parent, env := g.responseSchema(op)
		if parent == nil {
			return nil, nil, fmt.Errorf("%s %s has no JSON response schema", op.method, op.path)
		}
		f := g.norm(properties(parent)[rf.Field])
		if f == nil || f["type"] != "array" {
			return nil, nil, fmt.Errorf("%s %s has no array field %s", op.method, op.path, rf.Field)
		}
		rf.Op = &engine.Op{Method: op.method, Path: op.path, Envelope: env}
		item = g.norm(f["items"])
	}
	if t, _ := item["type"].(string); t == "object" || properties(item) != nil {
		if rf.Key == "" {
			return nil, nil, fmt.Errorf("elements are objects; set key")
		}
		if properties(item)[rf.Key] == nil {
			return nil, nil, fmt.Errorf("elements have no field %s", rf.Key)
		}
		return item, rf, nil
	}
	if rf.Key != "" {
		return nil, nil, fmt.Errorf("elements are plain values; leave key empty")
	}
	if ov.ID == "" {
		return nil, nil, fmt.Errorf("elements are plain values; set id to the create field that holds the value")
	}
	return map[string]any{"type": "object", "properties": map[string]any{ov.ID: item}}, rf, nil
}

// extraUpdates derives the attributes each extra update operation sets and
// records the operations on r. It returns the operations, for their path
// parameters.
func (g *Generator) extraUpdates(r *engine.Resource, ov *Override, d *deriver, readSchema map[string]any) ([]*operation, error) {
	var ops []*operation
	owner := map[string]string{}
	for _, u := range ov.Updates {
		op, err := g.operation(u.Op)
		if err != nil {
			return nil, fmt.Errorf("updates: %w", err)
		}
		clearOp, err := g.operation(u.Clear)
		if err != nil {
			return nil, fmt.Errorf("updates: clear: %w", err)
		}
		body := g.requestSchema(op)
		if body == nil {
			return nil, fmt.Errorf("updates: %s has no request body", u.Op)
		}
		ud := &deriver{g: g, ov: ov, used: d.used, hasUpdate: true}
		as, err := ud.attrs(sources{update: body, read: readSchema}, 0, false, "")
		if err != nil {
			return nil, err
		}
		x := engine.ExtraUpdate{Op: &engine.Op{Method: op.method, Path: op.path}}
		if _, env := g.responseSchema(op); env {
			x.Op.Envelope = true
		}
		if clearOp != nil {
			x.Clear = &engine.Op{Method: clearOp.method, Path: clearOp.path}
			ops = append(ops, clearOp)
		}
		for _, a := range as {
			if !a.InUpdate {
				// Only read returns it.
				continue
			}
			if a.WriteOnly {
				return nil, fmt.Errorf("updates: %s sets write-only %s; write_only is not supported there", u.Op, a.Name)
			}
			if prev, ok := owner[a.Name]; ok {
				return nil, fmt.Errorf("updates %s and %s both set %s", prev, u.Op, a.Name)
			}
			owner[a.Name] = u.Op
			// Removing the attribute from configuration must read as null so
			// that the clear operation runs: optional, never computed. With
			// no clear operation there is nothing to run, and the API's value
			// is kept: optional and computed.
			a.Required, a.Optional, a.Computed, a.ForceNew = false, true, u.Clear == "" && a.InResponse, false
			a.InCreate, a.InUpdate = false, false
			if i := slices.IndexFunc(r.Attrs, func(b *engine.Attr) bool { return b.Name == a.Name }); i >= 0 {
				a.InCreate = r.Attrs[i].InCreate
				if r.Attrs[i].Required {
					// Create needs it: it stays required.
					a.Required, a.Optional, a.Computed = true, false, false
				}
				r.Attrs[i] = a
			} else {
				r.Attrs = append(r.Attrs, a)
			}
			x.Attrs = append(x.Attrs, a.Name)
		}
		if len(x.Attrs) == 0 {
			return nil, fmt.Errorf("updates: %s sets no attributes", u.Op)
		}
		sort.Strings(x.Attrs)
		r.Updates = append(r.Updates, x)
		ops = append(ops, op)
	}
	return ops, nil
}

// ids wires the id attribute and parent path parameters, and sets Import.
// collection, when set, is the read_from operation: all its path
// parameters are parents.
func (g *Generator) ids(r *engine.Resource, ov *Override, create *operation, others []*operation, collection *operation) error {
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
	if ov.Singleton {
		if len(createParams) > 0 {
			return fmt.Errorf("singleton: %s has path parameters", create.path)
		}
		// The API's own id of a once-per-account object identifies nothing
		// the user can choose: the fixed id replaces it.
		r.Attrs = slices.DeleteFunc(r.Attrs, func(a *engine.Attr) bool { return a.Name == "id" })
		r.FixedID = strings.TrimPrefix(r.TypeName, "telnyx_")
		r.Attrs = append(r.Attrs, &engine.Attr{Name: "id", JSON: "id", Kind: engine.KindString, Computed: true, InResponse: true,
			Description: "Always " + r.FixedID + ": there is one per account."})
		r.IDAttr = "id"
		r.Import = []string{"id"}
		return nil
	}
	if ov.IDFromPath != "" {
		i := slices.Index(createParams, ov.IDFromPath)
		if i < 0 {
			return fmt.Errorf("id_from_path %s is not a parameter of %s", ov.IDFromPath, create.path)
		}
		id = parents[i]
		r.IDAttr = id.Name
		// Importing needs every path parameter, parents first.
		r.Import = nil
		for _, p := range parents {
			r.Import = append(r.Import, p.Name)
		}
		for _, op := range append(others, collection) {
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
	if field == "" && ov.ReadFrom != nil {
		field = ov.ReadFrom.Key
	}
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
	if ov.ReadFrom != nil && ov.ReadFrom.Key == "" {
		// A plain value is its own id: the user sets it, and a new value is
		// a new element.
		if !id.InCreate {
			return fmt.Errorf("%s is not in the create body", field)
		}
		id.Required, id.Optional, id.Computed, id.ForceNew = true, false, false, true
		id.InUpdate, id.InResponse = false, true
		id.Kind = engine.KindString
	} else if (id.InCreate || id.Required) && field != "id" {
		// The user names the object (an id override such as a name or a
		// phone number): the field is sent on create (by a hook when the
		// override makes a response field required), identifies the
		// object, and a new value is a new object.
		id.Required, id.Optional, id.Computed, id.ForceNew = true, false, false, true
		id.InUpdate = false
		id.Kind = engine.KindString
	} else {
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
	}
	r.IDAttr = id.Name
	for _, op := range others {
		if op == nil {
			continue
		}
		ps := pathParams(op.path)
		if len(ps) == 0 {
			return fmt.Errorf("%s %s has no id parameter", op.method, op.path)
		}
		// An operation on the collection, with the create path's parameters
		// and no more, takes the object in its body: all are parents.
		collectionOp := len(ps) == len(createParams) && slices.Equal(ps, createParams)
		for j, p := range ps {
			target := id
			if j < len(ps)-1 || collectionOp {
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
	if collection != nil {
		for j, p := range pathParams(collection.path) {
			if j >= len(parents) {
				return fmt.Errorf("%s %s has a parameter %s that create does not", collection.method, collection.path, p)
			}
			if !slices.Contains(parents[j].PathParams, p) {
				parents[j].PathParams = append(parents[j].PathParams, p)
			}
		}
	}
	for _, p := range parents {
		r.Import = append(r.Import, p.Name)
	}
	r.Import = append(r.Import, r.IDAttr)
	if ov.ReadFrom != nil {
		for _, n := range ov.ReadFrom.Match {
			a := r.Attr(n)
			if a == nil || a.Local || a.Kind != engine.KindString {
				return fmt.Errorf("read_from: match %s is not a string attribute of the API", n)
			}
			a.Required, a.Optional, a.Computed, a.ForceNew = true, false, false, true
			a.InUpdate = false
			r.ReadFrom.Match = append(r.ReadFrom.Match, n)
			r.Import = append(r.Import, n)
		}
	}
	return nil
}

// queryParams records on update and delete operations the attributes they
// take as query parameters.
func (g *Generator) queryParams(r *engine.Resource, ov *Override) error {
	type target struct {
		op  *engine.Op
		key string
	}
	var ts []target
	if r.Update != nil {
		ts = append(ts, target{r.Update, r.Update.Method + " " + r.Update.Path})
	}
	if r.Delete != nil {
		ts = append(ts, target{r.Delete, r.Delete.Method + " " + r.Delete.Path})
	}
	for _, u := range r.Updates {
		ts = append(ts, target{u.Op, u.Op.Method + " " + u.Op.Path})
	}
	for _, n := range ov.Query {
		a := r.Attr(n)
		if a == nil || a.Local || a.Kind != engine.KindString {
			return fmt.Errorf("query: %s is not a string attribute of the API", n)
		}
		used := false
		for _, t := range ts {
			op, err := g.operation(t.key)
			if err != nil {
				return err
			}
			for _, p := range g.params(op) {
				if p.in == "query" && p.name == a.JSON {
					t.op.Query = append(t.op.Query, a.JSON)
					used = true
				}
			}
		}
		if !used {
			return fmt.Errorf("query: no update or delete operation takes %s", a.JSON)
		}
	}
	return nil
}

// mergeSameShape merges object variants whose fields agree in type and
// differ only in enum values, in which fields are required and in fields
// some variants leave out, such as a provider choice that changes the
// accepted models. Enums are joined unless a variant accepts any value,
// and a field is required when every variant requires it.
func (g *Generator) mergeSameShape(vs []map[string]any) map[string]any {
	merged := map[string]map[string]any{}
	open := map[string]bool{} // a variant takes the field without an enum
	count := map[string]int{}
	for _, v := range vs {
		p := properties(v)
		if p == nil {
			return nil
		}
		for k, pv := range p {
			b := g.norm(pv)
			if b == nil || b["type"] == nil || b["type"] == "object" || b["type"] == "array" {
				return nil
			}
			count[k]++
			eb, hasEnum := b["enum"].([]any)
			if !hasEnum {
				open[k] = true
			}
			a, seen := merged[k]
			if !seen {
				c := map[string]any{}
				for x, y := range b {
					c[x] = y
				}
				merged[k] = c
				continue
			}
			if a["type"] != b["type"] {
				return nil
			}
			ea, _ := a["enum"].([]any)
			all := append([]any{}, ea...)
			for _, e := range eb {
				dup := false
				for _, x := range all {
					dup = dup || x == e
				}
				if !dup {
					all = append(all, e)
				}
			}
			a["enum"] = all
		}
	}
	props := map[string]any{}
	var req []any
	for k, a := range merged {
		if open[k] {
			delete(a, "enum")
		}
		props[k] = a
		all := count[k] == len(vs)
		for _, v := range vs {
			all = all && required(v, k)
		}
		if all {
			req = append(req, k)
		}
	}
	sort.Slice(req, func(i, j int) bool { return fmt.Sprint(req[i]) < fmt.Sprint(req[j]) })
	return map[string]any{"type": "object", "properties": props, "required": req}
}
