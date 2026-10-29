// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package apisync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var slugChars = regexp.MustCompile(`[^a-z0-9]+`)

// TagSlug turns a tag such as "Usage Reports (BETA)" into "usage-reports-beta".
func TagSlug(tag string) string {
	return strings.Trim(slugChars.ReplaceAllString(strings.ToLower(tag), "-"), "-")
}

// SplitByTag returns one self-contained OpenAPI document per tag, keyed by
// tag slug. Each holds only that tag's operations and the components they
// reference, so an agent can read one product area without the full spec.
func SplitByTag(s *Spec) map[string]map[string]any {
	type bucket struct {
		paths map[string]any
		nodes []any
	}
	byTag := map[string]*bucket{}
	paths := s.Paths()
	for _, op := range s.Operations() {
		slug := TagSlug(op.Tag())
		b := byTag[slug]
		if b == nil {
			b = &bucket{paths: map[string]any{}}
			byTag[slug] = b
		}
		item, _ := b.paths[op.Path].(map[string]any)
		if item == nil {
			item = map[string]any{}
			// Path-level parameters apply to every method on the path.
			if params, ok := asMap(paths[op.Path])["parameters"]; ok {
				item["parameters"] = params
				b.nodes = append(b.nodes, params)
			}
			b.paths[op.Path] = item
		}
		item[op.Method] = op.Node
		b.nodes = append(b.nodes, op.Node)
	}

	out := map[string]map[string]any{}
	comps := asMap(s.Doc["components"])
	for slug, b := range byTag {
		components := map[string]any{}
		if sec, ok := comps["securitySchemes"]; ok {
			components["securitySchemes"] = sec
		}
		for _, ref := range s.Refs(b.nodes...) {
			parts := strings.SplitN(strings.TrimPrefix(ref, "#/components/"), "/", 2)
			if len(parts) != 2 || !strings.HasPrefix(ref, "#/components/") {
				continue
			}
			kind := asMap(components[parts[0]])
			components[parts[0]] = kind
			v, _ := s.Component(ref)
			kind[parts[1]] = v
		}
		doc := map[string]any{
			"openapi":    s.Doc["openapi"],
			"info":       s.Doc["info"],
			"servers":    s.Doc["servers"],
			"paths":      b.paths,
			"components": components,
		}
		if sec, ok := s.Doc["security"]; ok {
			doc["security"] = sec
		}
		out[slug] = doc
	}
	return out
}

// WriteSplit writes SplitByTag output to dir as <slug>.json and removes files
// for tags that no longer exist.
func WriteSplit(s *Spec, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	docs := SplitByTag(s)
	for slug, doc := range docs {
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, slug+".json"), append(data, '\n'), 0o644); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			if _, ok := docs[strings.TrimSuffix(name, ".json")]; !ok {
				if err := os.Remove(filepath.Join(dir, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
