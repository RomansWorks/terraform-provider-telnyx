// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// Inexplicit number orders take each ordering group's count_requested as a
// string and return it as a number. The read hook returns it as a string.
var inexplicitNumberOrderHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		groups, ok := obj["ordering_groups"].([]any)
		if !ok {
			return obj, nil
		}
		out := make(engine.Object, len(obj))
		for k, v := range obj {
			out[k] = v
		}
		reshaped := make([]any, len(groups))
		for i, g := range groups {
			m, ok := g.(map[string]any)
			if !ok {
				reshaped[i] = g
				continue
			}
			c := make(map[string]any, len(m))
			for k, v := range m {
				c[k] = v
			}
			if n, ok := c["count_requested"].(json.Number); ok {
				c["count_requested"] = n.String()
			}
			reshaped[i] = c
		}
		out["ordering_groups"] = reshaped
		return out, nil
	},
}
