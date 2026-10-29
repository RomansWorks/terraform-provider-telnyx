// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

// Command codegen derives resource descriptors from the pinned Telnyx
// OpenAPI spec and the override files, and writes them as Go. Run it with
// `make generate`.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/codegen"
)

func main() {
	spec := flag.String("spec", "upstream/telnyx/spec3.json", "OpenAPI spec")
	overrides := flag.String("overrides", "generator/overrides", "override files")
	out := flag.String("out", "internal/provider/generated", "output directory")
	flag.Parse()

	s, err := apisync.LoadSpec(*spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen:", err)
		os.Exit(1)
	}
	rs, err := codegen.Run(&codegen.Generator{Spec: s}, *overrides, *out, "generated")
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen:", err)
		os.Exit(1)
	}
	fmt.Printf("codegen: %d resources in %s\n", len(rs), *out)
}
