// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

// Package codegen derives resource descriptors from the pinned Telnyx
// OpenAPI spec and per-resource override files, and writes them as Go.
package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Override is one file in generator/overrides: it names a resource's
// operations and adjusts only what the spec cannot say.
type Override struct {
	// Type is the Terraform type name, such as telnyx_ai_assistant.
	Type        string `yaml:"type"`
	Description string `yaml:"description"`

	// Operations as "METHOD /path".
	Create string `yaml:"create"`
	Read   string `yaml:"read"`
	Update string `yaml:"update"`
	Delete string `yaml:"delete"`

	// ID is the response field holding the object's id. Default "id".
	ID string `yaml:"id"`
	// IDFromPath names the path parameter that identifies the object when
	// the object has no id of its own (settings of another object).
	IDFromPath string `yaml:"id_from_path"`
	// PathParams renames path parameters that become attributes, such as
	// {"id": "phone_number_id"}.
	PathParams map[string]string `yaml:"path_params"`

	// DataSource is true unless set to false.
	DataSource    *bool      `yaml:"data_source"`
	Timeouts      bool       `yaml:"timeouts"`
	DeleteWarning string     `yaml:"delete_warning"`
	ExactlyOne    [][]string `yaml:"exactly_one"`

	// Attributes adjusts derived attributes by JSON path ("a.b.c").
	Attributes map[string]AttrOverride `yaml:"attributes"`
	// Add declares attributes that are not in the spec. They are local:
	// hooks read them and the engine never sends them.
	Add []AddAttr `yaml:"add"`

	file string
}

// AttrOverride adjusts one derived attribute.
type AttrOverride struct {
	Name        string `yaml:"name"`
	Skip        bool   `yaml:"skip"`
	JSON        bool   `yaml:"json"`
	Set         bool   `yaml:"set"`
	Sensitive   bool   `yaml:"sensitive"`
	ForceNew    *bool  `yaml:"force_new"`
	Required    *bool  `yaml:"required"`
	Computed    *bool  `yaml:"computed"`
	Description string `yaml:"description"`
	// Returned says a hook puts the value into the read response although
	// the spec does not list it there, so drift is detected.
	Returned bool `yaml:"returned"`
}

// AddAttr is a local attribute.
type AddAttr struct {
	Name        string `yaml:"name"`
	Kind        string `yaml:"kind"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Optional    bool   `yaml:"optional"`
	Computed    bool   `yaml:"computed"`
	Sensitive   bool   `yaml:"sensitive"`
	ForceNew    bool   `yaml:"force_new"`
}

var (
	opPattern   = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE) /\S*$`)
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// LoadOverrides reads every *.yaml file in dir, sorted by name.
func LoadOverrides(dir string) ([]*Override, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []*Override
	seen := map[string]string{}
	for _, f := range files {
		o, err := LoadOverride(f)
		if err != nil {
			return nil, err
		}
		if prev, ok := seen[o.Type]; ok {
			return nil, fmt.Errorf("%s: type %s is also defined in %s", f, o.Type, prev)
		}
		seen[o.Type] = f
		out = append(out, o)
	}
	return out, nil
}

// LoadOverride reads and validates one override file. Unknown keys are
// errors, so a typo cannot silently do nothing.
func LoadOverride(path string) (*Override, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var o Override
	if err := dec.Decode(&o); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	o.file = path
	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &o, nil
}

func (o *Override) validate() error {
	if !strings.HasPrefix(o.Type, "telnyx_") || !namePattern.MatchString(o.Type) {
		return fmt.Errorf("type %q must be telnyx_<snake_case>", o.Type)
	}
	if o.Description == "" {
		return fmt.Errorf("description is required")
	}
	if o.Create == "" || o.Read == "" {
		return fmt.Errorf("create and read are required")
	}
	for _, op := range []string{o.Create, o.Read, o.Update, o.Delete} {
		if op != "" && !opPattern.MatchString(op) {
			return fmt.Errorf("operation %q must be \"METHOD /path\"", op)
		}
	}
	for _, a := range o.Add {
		if !namePattern.MatchString(a.Name) {
			return fmt.Errorf("added attribute %q must be snake_case", a.Name)
		}
		if _, ok := addKinds[a.Kind]; !ok {
			return fmt.Errorf("added attribute %s: kind must be string, int, float, bool or json", a.Name)
		}
		if a.Required == (a.Optional || a.Computed) {
			return fmt.Errorf("added attribute %s must be either required, or optional and/or computed", a.Name)
		}
	}
	for p, a := range o.Attributes {
		if a.Name != "" && !namePattern.MatchString(a.Name) {
			return fmt.Errorf("attribute %s: name %q must be snake_case", p, a.Name)
		}
	}
	return nil
}

// dataSource reports whether the resource also gets a data source.
func (o *Override) dataSource() bool { return o.DataSource == nil || *o.DataSource }
