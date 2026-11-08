// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

// Command examplecheck validates every example in examples/ with a Terraform
// or OpenTofu CLI against a locally built provider. It needs no Telnyx
// credentials: validation never calls the API.
//
//	go build -o build/terraform-provider-telnyx .
//	go run ./cmd/examplecheck -cli tofu -provider build/terraform-provider-telnyx -kinds resources,data-sources,ephemeral-resources
//
// Each example directory is checked on its own, together with
// examples/provider/provider.tf. An example that refers to an object it does
// not declare (the SIM card an action enables, say) gets that object from the
// example for its resource type, under the referenced name; when that type
// has no example, a stub with placeholder values for its required
// attributes, built from the provider schema. Variables come from whichever
// example declares them.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	refRe = regexp.MustCompile(`(?:^|[^.\w])(telnyx_\w+)\.(\w+)\.`)
	varRe = regexp.MustCompile(`\bvar\.(\w+)`)
	// typeRe finds the resource and data source types an example uses.
	typeRe = regexp.MustCompile(`(?m)^(?:resource|data|ephemeral) "(\w+)"`)
)

func main() {
	cli := flag.String("cli", "terraform", "terraform or tofu, or a path to either")
	provider := flag.String("provider", "", "path to a built terraform-provider-telnyx binary")
	kinds := flag.String("kinds", "resources,data-sources,ephemeral-resources,actions", "comma-separated example folders to check")
	root := flag.String("examples", "examples", "the examples directory")
	flag.Parse()
	if *provider == "" {
		fmt.Fprintln(os.Stderr, "examplecheck: -provider is required")
		os.Exit(2)
	}
	n, failed, err := run(*cli, *provider, *root, strings.Split(*kinds, ","))
	if err != nil {
		fmt.Fprintln(os.Stderr, "examplecheck:", err)
		os.Exit(1)
	}
	fmt.Printf("%d examples checked with %s, %d failed.\n", n, *cli, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func run(cli, provider, root string, kinds []string) (int, int, error) {
	work, err := os.MkdirTemp("", "examplecheck")
	if err != nil {
		return 0, 0, err
	}
	defer os.RemoveAll(work)

	provider, err = filepath.Abs(provider)
	if err != nil {
		return 0, 0, err
	}
	plugins := filepath.Join(work, "plugins")
	if err := os.MkdirAll(plugins, 0o755); err != nil {
		return 0, 0, err
	}
	if err := os.Symlink(provider, filepath.Join(plugins, "terraform-provider-telnyx")); err != nil {
		return 0, 0, err
	}
	// Terraform and OpenTofu read the same CLI configuration format; they
	// differ only in the registry host a bare source address expands to.
	rc := fmt.Sprintf(`provider_installation {
  dev_overrides {
    "registry.terraform.io/romansworks/telnyx" = %[1]q
    "registry.opentofu.org/romansworks/telnyx" = %[1]q
  }
  direct {}
}
`, plugins)
	rcPath := filepath.Join(work, "cli.rc")
	if err := os.WriteFile(rcPath, []byte(rc), 0o644); err != nil {
		return 0, 0, err
	}
	env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+rcPath, "TF_IN_AUTOMATION=1", "CHECKPOINT_DISABLE=1", "TELNYX_API_KEY=")

	providerTF, err := os.ReadFile(filepath.Join(root, "provider", "provider.tf"))
	if err != nil {
		return 0, 0, err
	}
	schemas, err := loadSchemas(cli, env, work, providerTF)
	if err != nil {
		return 0, 0, err
	}
	ex := &library{root: root, schemas: schemas}

	checked, failed := 0, 0
	for _, kind := range kinds {
		dirs, _ := filepath.Glob(filepath.Join(root, kind, "*"))
		sort.Strings(dirs)
		for _, dir := range dirs {
			files, _ := filepath.Glob(filepath.Join(dir, "*.tf"))
			if len(files) == 0 {
				continue // only an import.sh
			}
			caseDir := filepath.Join(work, "case")
			_ = os.RemoveAll(caseDir)
			if err := os.MkdirAll(caseDir, 0o755); err != nil {
				return checked, failed, err
			}
			srcs := map[string][]byte{"provider.tf": providerTF}
			for _, f := range files {
				b, err := os.ReadFile(f)
				if err != nil {
					return checked, failed, err
				}
				srcs[filepath.Base(f)] = b
			}
			ex.complete(srcs)
			for name, b := range srcs {
				if err := os.WriteFile(filepath.Join(caseDir, name), b, 0o644); err != nil {
					return checked, failed, err
				}
			}
			checked++
			out, err := validate(cli, env, caseDir, usesOtherProviders(joinedBytes(srcs)))
			if err != nil {
				failed++
				fmt.Printf("FAIL %s/%s\n%s\n", kind, filepath.Base(dir), indent(stripOverrideWarning(out)))
			}
		}
	}
	return checked, failed, nil
}

// validate runs validate in dir, after init when an example needs providers
// other than telnyx (dev_overrides cover telnyx without init).
func validate(cli string, env []string, dir string, needsInit bool) ([]byte, error) {
	if needsInit {
		cmd := exec.Command(cli, "init", "-backend=false", "-no-color", "-input=false")
		cmd.Dir, cmd.Env = dir, env
		if out, err := cmd.CombinedOutput(); err != nil {
			return out, err
		}
	}
	cmd := exec.Command(cli, "validate", "-no-color")
	cmd.Dir, cmd.Env = dir, env
	return cmd.CombinedOutput()
}

// loadSchemas returns the provider's resource schemas, keyed by type.
func loadSchemas(cli string, env []string, work string, providerTF []byte) (map[string]block, error) {
	dir := filepath.Join(work, "schema")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "provider.tf"), providerTF, 0o644); err != nil {
		return nil, err
	}
	cmd := exec.Command(cli, "providers", "schema", "-json")
	cmd.Dir, cmd.Env = dir, env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s providers schema: %v\n%s", cli, err, stderr.String())
	}
	var doc struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block block `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, err
	}
	res := map[string]block{}
	for addr, p := range doc.ProviderSchemas {
		if !strings.HasSuffix(addr, "/romansworks/telnyx") {
			continue
		}
		for name, r := range p.ResourceSchemas {
			res[name] = r.Block
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("%s providers schema returned no telnyx resources", cli)
	}
	return res, nil
}

type block struct {
	Attributes map[string]struct {
		Type       json.RawMessage `json:"type"`
		NestedType *struct {
			Attributes map[string]json.RawMessage `json:"attributes"`
		} `json:"nested_type"`
		Required bool `json:"required"`
	} `json:"attributes"`
}

type library struct {
	root    string
	schemas map[string]block
}

// complete adds the resources and variables srcs refers to but does not
// declare, until nothing is missing.
func (l *library) complete(srcs map[string][]byte) {
	for range 20 {
		all := joined(srcs)
		added := false
		for _, m := range refRe.FindAllStringSubmatch(all, -1) {
			typ, label := m[1], m[2]
			name := "stub_" + typ + "_" + label + ".tf"
			if _, ok := srcs[name]; ok || declares(all, "resource", typ, label) {
				continue
			}
			srcs[name] = l.resource(typ, label)
			added = true
		}
		for _, m := range varRe.FindAllStringSubmatch(all, -1) {
			name := "stub_var_" + m[1] + ".tf"
			if _, ok := srcs[name]; ok || declares(all, "variable", m[1]) {
				continue
			}
			srcs[name] = l.variable(m[1])
			added = true
		}
		if !added {
			return
		}
	}
}

// resource returns a block for typ named label: the one in typ's example with
// that name, else the first one there, else a stub from the schema.
func (l *library) resource(typ, label string) []byte {
	b, _ := os.ReadFile(filepath.Join(l.root, "resources", typ, "resource.tf"))
	src := string(b)
	if blk := extract(src, fmt.Sprintf("resource %q %q", typ, label)); blk != "" {
		return []byte(blk)
	}
	if blk := extract(src, fmt.Sprintf("resource %q ", typ)); blk != "" {
		_, rest, _ := strings.Cut(blk, "{")
		return []byte(fmt.Sprintf("resource %q %q {%s", typ, label, rest))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "resource %q %q {\n", typ, label)
	s, ok := l.schemas[typ]
	if ok {
		names := make([]string, 0, len(s.Attributes))
		for n, a := range s.Attributes {
			if a.Required {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			a := s.Attributes[n]
			v := "{}"
			if a.NestedType == nil {
				v = placeholder(a.Type)
			}
			fmt.Fprintf(&sb, "  %s = %s\n", n, v)
		}
	}
	sb.WriteString("}\n")
	return []byte(sb.String())
}

// variable returns the first declaration of name in any example.
func (l *library) variable(name string) []byte {
	var found string
	_ = filepath.WalkDir(l.root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found != "" || d.IsDir() || !strings.HasSuffix(p, ".tf") {
			return nil
		}
		b, _ := os.ReadFile(p)
		found = extract(string(b), fmt.Sprintf("variable %q", name))
		return nil
	})
	if found == "" {
		found = fmt.Sprintf("variable %q {}\n", name)
	}
	return []byte(found)
}

func placeholder(t json.RawMessage) string {
	var s string
	if json.Unmarshal(t, &s) == nil {
		switch s {
		case "string":
			return `"placeholder"`
		case "number":
			return "1"
		case "bool":
			return "true"
		}
		return "null"
	}
	var parts []json.RawMessage
	if json.Unmarshal(t, &parts) == nil && len(parts) > 0 {
		var kind string
		_ = json.Unmarshal(parts[0], &kind)
		switch kind {
		case "list", "set":
			return "[]"
		case "map", "object":
			return "{}"
		}
	}
	return "null"
}

// extract returns the top-level block whose header starts with prefix.
func extract(src, prefix string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(line), "{}") {
			return line + "\n"
		}
		for j := i + 1; j < len(lines); j++ {
			if lines[j] == "}" {
				return strings.Join(lines[i:j+1], "\n") + "\n"
			}
		}
	}
	return ""
}

func declares(src string, kind string, names ...string) bool {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = fmt.Sprintf("%q", n)
	}
	return regexp.MustCompile(`(?m)^` + kind + ` ` + regexp.QuoteMeta(strings.Join(quoted, " ")) + `\s*\{`).MatchString(src)
}

// usesOtherProviders reports whether src uses a provider other than telnyx
// and the built-in one, such as the AWS provider for storage buckets.
func usesOtherProviders(src []byte) bool {
	for _, m := range typeRe.FindAllSubmatch(src, -1) {
		t := string(m[1])
		if !strings.HasPrefix(t, "telnyx_") && !strings.HasPrefix(t, "terraform_") {
			return true
		}
	}
	return false
}

// joinedBytes is joined for the example's own files, without provider.tf.
func joinedBytes(srcs map[string][]byte) []byte {
	var b []byte
	for name, src := range srcs {
		if name != "provider.tf" {
			b = append(b, src...)
		}
	}
	return b
}

func joined(srcs map[string][]byte) string {
	var sb strings.Builder
	for _, b := range srcs {
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// stripOverrideWarning drops the dev_overrides warning both CLIs print.
func stripOverrideWarning(out []byte) string {
	s := string(out)
	if i := strings.Index(s, "Warning: Provider development overrides are in effect"); i >= 0 {
		if j := strings.Index(s[i:], "published\nreleases."); j >= 0 {
			s = s[:i] + s[i+j+len("published\nreleases."):]
		}
	}
	return strings.TrimSpace(s)
}

func indent(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}
