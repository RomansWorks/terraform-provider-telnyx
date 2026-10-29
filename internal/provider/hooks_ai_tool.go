// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// aiToolVariants are the attributes that hold a tool's settings, named after
// the tool types.
var aiToolVariants = []string{"function", "retrieval", "handoff", "invite", "webhook", "pay", "client_side_tool", "update_dynamic_variables", "vision"}

// The API takes a tool's settings in an attribute named by its type (for
// example webhook) and returns them as tool_definition. Copying them back
// lets Terraform detect drift and import tools.
var aiToolHooks = engine.Hooks{
	ModifyPlan: func(_ context.Context, plan, state engine.Values) ([]string, diag.Diagnostics) {
		var diags diag.Diagnostics
		var typ string
		if t := plan["type"]; !t.IsKnown() || t.IsNull() || t.As(&typ) != nil {
			return nil, nil
		}
		for _, v := range aiToolVariants {
			p := plan[v]
			if v != typ && p.IsKnown() && !p.IsNull() {
				diags.AddAttributeError(pathRoot(v), "Tool settings do not match its type",
					fmt.Sprintf("type is %q, so set %s instead of %s.", typ, typ, v))
			}
		}
		// The API does not change a tool's type in place.
		if old, ok := state["type"]; ok && !old.IsNull() && !old.Equal(plan["type"]) {
			return []string{"type"}, diags
		}
		return nil, diags
	},
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
