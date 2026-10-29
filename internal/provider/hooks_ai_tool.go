// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// The API takes a tool's settings in an attribute named by its type (for
// example webhook) and returns them as tool_definition. Copying them back
// lets Terraform detect drift and import tools.
var aiToolHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		t, _ := obj["type"].(string)
		def, ok := obj["tool_definition"].(map[string]any)
		if t != "" && ok {
			if _, set := obj[t]; !set {
				obj[t] = def
			}
		}
		return obj, nil
	},
}
