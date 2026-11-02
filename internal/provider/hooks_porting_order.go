// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// A porting order is created from a list of phone numbers, but read back
// with an object per number. The objects are reduced to their numbers so
// the configured set can be compared with what the order holds.
var portingOrderHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		items, ok := obj["phone_numbers"].([]any)
		if !ok {
			return obj, nil
		}
		numbers := make([]any, 0, len(items))
		for _, it := range items {
			switch v := it.(type) {
			case map[string]any:
				if n, ok := v["phone_number"]; ok {
					numbers = append(numbers, n)
				}
			default:
				numbers = append(numbers, v)
			}
		}
		out := make(engine.Object, len(obj))
		for k, v := range obj {
			out[k] = v
		}
		out["phone_numbers"] = numbers
		return out, nil
	},
}
