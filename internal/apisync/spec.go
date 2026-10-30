// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

// Package apisync pins the upstream Telnyx OpenAPI spec and derives the
// agent-sized views of it that the rest of the repository works from: a
// per-tag split, a resource catalog and the operation coverage file.
package apisync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Methods are the HTTP methods treated as operations, in display order.
var Methods = []string{"get", "post", "put", "patch", "delete"}

// Spec is a parsed OpenAPI document kept as generic JSON so that nothing in
// the upstream file is lost or reinterpreted.
type Spec struct {
	Doc map[string]any
}

// Operation is one method on one path.
type Operation struct {
	Method      string
	Path        string
	OperationID string
	Tags        []string
	Summary     string
	Deprecated  bool
	Beta        bool
	Hidden      bool
	Node        map[string]any
}

// Key identifies an operation as "METHOD /path". OperationIDs are not unique
// in the Telnyx spec (aliased paths share them), so this is the stable key.
func (o Operation) Key() string { return strings.ToUpper(o.Method) + " " + o.Path }

// Tag returns the operation's first tag, or "untagged".
func (o Operation) Tag() string {
	if len(o.Tags) == 0 {
		return "untagged"
	}
	return o.Tags[0]
}

// LoadSpec reads an OpenAPI JSON document.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if _, ok := doc["paths"].(map[string]any); !ok {
		return nil, fmt.Errorf("%s has no paths object", path)
	}
	return &Spec{Doc: doc}, nil
}

// Paths returns the paths object.
func (s *Spec) Paths() map[string]any { return asMap(s.Doc["paths"]) }

// Operations returns every operation sorted by path, then method order.
func (s *Spec) Operations() []Operation {
	paths := s.Paths()
	keys := sortedKeys(paths)
	var ops []Operation
	for _, p := range keys {
		item := asMap(paths[p])
		for _, m := range Methods {
			n, ok := item[m].(map[string]any)
			if !ok {
				continue
			}
			op := Operation{Method: m, Path: p, Node: n}
			op.OperationID, _ = n["operationId"].(string)
			op.Summary, _ = n["summary"].(string)
			op.Deprecated, _ = n["deprecated"].(bool)
			op.Beta, _ = n["x-beta"].(bool)
			op.Hidden, _ = n["x-hidden"].(bool)
			for _, t := range asSlice(n["tags"]) {
				if ts, ok := t.(string); ok {
					op.Tags = append(op.Tags, ts)
				}
			}
			ops = append(ops, op)
		}
	}
	return ops
}

// Methods returns the set of methods defined on a path.
func (s *Spec) MethodsOn(path string) map[string]bool {
	out := map[string]bool{}
	item := asMap(s.Paths()[path])
	for _, m := range Methods {
		if _, ok := item[m].(map[string]any); ok {
			out[m] = true
		}
	}
	return out
}

// Component resolves a local "#/components/..." reference.
func (s *Spec) Component(ref string) (any, bool) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, false
	}
	var cur any = s.Doc
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		if cur, ok = m[part]; !ok {
			return nil, false
		}
	}
	return cur, true
}

// Refs returns every local reference reachable from the given nodes,
// following references transitively.
func (s *Spec) Refs(nodes ...any) []string {
	seen := map[string]bool{}
	var walk func(any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if r, ok := v["$ref"].(string); ok && strings.HasPrefix(r, "#/") && !seen[r] {
				seen[r] = true
				if target, ok := s.Component(r); ok {
					walk(target)
				}
			}
			for _, k := range sortedKeys(v) {
				walk(v[k])
			}
		case []any:
			for _, e := range v {
				walk(e)
			}
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	out := make([]string, 0, len(seen))
	for r := range seen {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// Hash returns a short, stable digest of the given nodes together with every
// component they reference, so a change anywhere in a resource's request or
// response shape changes the hash.
func (s *Spec) Hash(nodes ...any) string {
	refs := s.Refs(nodes...)
	resolved := make([]any, 0, len(refs))
	for _, r := range refs {
		c, _ := s.Component(r)
		resolved = append(resolved, map[string]any{"ref": r, "value": c})
	}
	data, _ := json.Marshal(map[string]any{"nodes": nodes, "refs": resolved})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
