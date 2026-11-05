// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// A DIR takes its call reasons as strings and answers with objects that
// also carry when each reason was added. The read hook keeps the strings,
// so call_reasons detects drift.
var displayIdentityRecordHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		list, ok := obj["call_reasons"].([]any)
		if !ok {
			return obj, nil
		}
		reasons := make([]any, 0, len(list))
		for _, el := range list {
			if m, ok := el.(map[string]any); ok {
				reasons = append(reasons, m["reason"])
				continue
			}
			reasons = append(reasons, el)
		}
		out := make(engine.Object, len(obj))
		for k, v := range obj {
			out[k] = v
		}
		out["call_reasons"] = reasons
		return out, nil
	},
}
