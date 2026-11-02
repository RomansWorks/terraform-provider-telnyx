// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// TeXML queues take form fields named like Twilio's (MaxSize) but answer in
// snake_case (max_size). The read hook copies the answer to the request
// name so max_size detects drift.
var texmlQueueHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		v, ok := obj["max_size"]
		if !ok {
			return obj, nil
		}
		out := make(engine.Object, len(obj)+1)
		for k, val := range obj {
			out[k] = val
		}
		out["MaxSize"] = v
		return out, nil
	},
}
