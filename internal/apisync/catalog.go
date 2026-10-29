// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package apisync

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Resource is a candidate Terraform resource found in the spec. The catalog
// is a starting point: generator overrides decide the final name and shape.
type Resource struct {
	Name           string            `yaml:"name"`
	Tag            string            `yaml:"tag"`
	Kind           string            `yaml:"kind"`
	CollectionPath string            `yaml:"collection_path,omitempty"`
	ItemPath       string            `yaml:"item_path"`
	IDParam        string            `yaml:"id_param,omitempty"`
	Operations     map[string]string `yaml:"operations"`
	Beta           bool              `yaml:"beta,omitempty"`
	SchemaHash     string            `yaml:"schema_hash"`
}

// Resource kinds.
const (
	KindCollection = "collection" // create on a collection, read/update/delete on an item
	KindSingleton  = "singleton"  // an object that always exists and is only read and updated
	KindExisting   = "existing"   // an object created elsewhere (an order, a port) and then managed by id
)

var (
	pathParam     = regexp.MustCompile(`^\{([^}]+)\}$`)
	versionPrefix = regexp.MustCompile(`^v\d+$`)
)

// DetectResources finds collection/item path pairs with a CRUD lifecycle,
// singleton settings objects, and objects created elsewhere that are then
// read and updated by id.
func DetectResources(s *Spec) []Resource {
	paths := sortedKeys(s.Paths())
	var out []Resource
	claimed := map[string]bool{}
	for _, p := range paths {
		last := lastSegment(p)
		if pathParam.MatchString(last) {
			continue
		}
		col := s.MethodsOn(p)
		item := itemPath(paths, p)

		if col["post"] && item != "" {
			im := s.MethodsOn(item)
			update := firstOf(im, "patch", "put", "post")
			if im["get"] && (update != "" || im["delete"]) {
				r := Resource{
					Tag:            s.opTag("post", p),
					Kind:           KindCollection,
					CollectionPath: p,
					ItemPath:       item,
					IDParam:        pathParam.FindStringSubmatch(lastSegment(item))[1],
					Operations: map[string]string{
						"create": "POST " + p,
						"read":   "GET " + item,
					},
				}
				if update != "" {
					r.Operations["update"] = strings.ToUpper(update) + " " + item
				}
				if im["delete"] {
					r.Operations["delete"] = "DELETE " + item
				}
				if col["get"] {
					r.Operations["list"] = "GET " + p
				}
				out = append(out, r)
				claimed[item] = true
				continue
			}
		}

		if col["get"] && !col["post"] {
			if update := firstOf(col, "patch", "put"); update != "" {
				out = append(out, Resource{
					Tag:      s.opTag("get", p),
					Kind:     KindSingleton,
					ItemPath: p,
					Operations: map[string]string{
						"read":   "GET " + p,
						"update": strings.ToUpper(update) + " " + p,
					},
				})
			}
		}
	}

	for _, p := range paths {
		m := pathParam.FindStringSubmatch(lastSegment(p))
		if m == nil || claimed[p] {
			continue
		}
		im := s.MethodsOn(p)
		update := firstOf(im, "patch", "put")
		if !im["get"] || update == "" {
			continue
		}
		r := Resource{
			Tag:      s.opTag("get", p),
			Kind:     KindExisting,
			ItemPath: p,
			IDParam:  m[1],
			Operations: map[string]string{
				"read":   "GET " + p,
				"update": strings.ToUpper(update) + " " + p,
			},
		}
		if im["delete"] {
			r.Operations["delete"] = "DELETE " + p
		}
		out = append(out, r)
	}

	ops := map[string]Operation{}
	for _, o := range s.Operations() {
		ops[o.Key()] = o
	}
	for i := range out {
		r := &out[i]
		var nodes []any
		for _, k := range sortedKeys(r.Operations) {
			o := ops[r.Operations[k]]
			nodes = append(nodes, o.Node)
			r.Beta = r.Beta || o.Beta
		}
		r.SchemaHash = s.Hash(nodes...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ItemPath < out[j].ItemPath })
	// Aliased paths (the spec has several) produce the same name; number the
	// later ones so every catalog entry stays addressable.
	used := map[string]int{}
	for i := range out {
		name := suggestName(&out[i])
		used[name]++
		if n := used[name]; n > 1 {
			name = fmt.Sprintf("%s_%d", name, n)
		}
		out[i].Name = name
	}
	return out
}
func (s *Spec) opTag(method, path string) string {
	n := asMap(asMap(s.Paths()[path])[method])
	for _, t := range asSlice(n["tags"]) {
		if ts, ok := t.(string); ok {
			return ts
		}
	}
	return "untagged"
}

func itemPath(sorted []string, collection string) string {
	depth := strings.Count(collection, "/") + 1
	for _, q := range sorted {
		if strings.HasPrefix(q, collection+"/{") && strings.Count(q, "/") == depth {
			return q
		}
	}
	return ""
}

func firstOf(set map[string]bool, methods ...string) string {
	for _, m := range methods {
		if set[m] {
			return m
		}
	}
	return ""
}

func lastSegment(p string) string {
	i := strings.LastIndex(p, "/")
	return p[i+1:]
}

// staticSegments returns the path segments that are not parameters, with the
// version prefixes some Telnyx paths carry removed.
func staticSegments(p string) []string {
	var out []string
	for _, seg := range strings.Split(strings.Trim(p, "/"), "/") {
		if seg == "" || pathParam.MatchString(seg) || versionPrefix.MatchString(seg) {
			continue
		}
		out = append(out, seg)
	}
	return out
}

// suggestName derives a Terraform type name from the item path's static
// segments: /ai/missions/{id}/tools/{id} becomes telnyx_ai_mission_tool.
func suggestName(r *Resource) string {
	segs := staticSegments(r.ItemPath)
	if len(segs) == 0 {
		return "telnyx_unnamed"
	}
	parts := make([]string, 0, len(segs))
	for i, seg := range segs {
		if i < len(segs)-1 || r.Kind != KindSingleton {
			seg = singular(seg)
		}
		parts = append(parts, seg)
	}
	return "telnyx_" + snake(strings.Join(parts, "_"))
}

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)
var camel = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func snake(s string) string {
	s = camel.ReplaceAllString(s, "${1}_${2}")
	s = nonWord.ReplaceAllString(strings.ToLower(s), "_")
	return strings.Trim(s, "_")
}

func singular(s string) string {
	switch {
	case strings.HasSuffix(s, "ies"):
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(s, "sses"), strings.HasSuffix(s, "xes"), strings.HasSuffix(s, "ches"):
		return s[:len(s)-2]
	case strings.HasSuffix(s, "ss"), strings.HasSuffix(s, "us"), strings.HasSuffix(s, "fs"):
		return s
	case strings.HasSuffix(s, "s"):
		return s[:len(s)-1]
	}
	return s
}
