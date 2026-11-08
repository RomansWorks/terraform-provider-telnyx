// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// Email domains take inbound_enabled but answer with inbound.enabled. The
// read hook copies it back so inbound_enabled detects drift and imports.
var emailDomainHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		in, _ := obj["inbound"].(map[string]any)
		v, ok := in["enabled"].(bool)
		if !ok {
			return obj, nil
		}
		return withField(obj, "inbound_enabled", v), nil
	},
}
