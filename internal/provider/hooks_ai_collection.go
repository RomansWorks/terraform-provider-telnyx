// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// A collection's retrieval settings are sent to its settings endpoint as
// retrieval but read back inside settings. The read hook copies them to
// retrieval so drift is detected.
var aiCollectionHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		settings, ok := obj["settings"].(map[string]any)
		if !ok {
			return obj, nil
		}
		out := make(engine.Object, len(obj)+1)
		for k, v := range obj {
			out[k] = v
		}
		out["retrieval"] = settings["retrieval"]
		return out, nil
	},
}
