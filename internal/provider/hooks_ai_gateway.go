// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"math/big"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// gatewayServerFields change on every update; the rest come from
// configuration.
var gatewayServerFields = []string{"version", "updated_at"}

// AI gateway token groups and keys change only with an If-Match header that
// holds the version the change is based on, quoted like the ETag the API
// sends. A change made outside Terraform since the last refresh fails with
// 412 instead of being overwritten. Each update raises the version, so the
// plan leaves it unknown when anything else changes.
var aiGatewayHooks = engine.Hooks{
	ModifyPlan: func(_ context.Context, plan, state engine.Values) ([]string, diag.Diagnostics) {
		if state == nil {
			return nil, nil
		}
		changed := false
		for name, v := range plan {
			if slices.Contains(gatewayServerFields, name) || !v.IsKnown() {
				continue
			}
			if old, ok := state[name]; ok && !old.Equal(v) {
				changed = true
			}
		}
		if changed {
			for _, name := range gatewayServerFields {
				if v, ok := plan[name]; ok {
					plan[name] = tftypes.NewValue(v.Type(), tftypes.UnknownValue)
				}
			}
		}
		return nil, nil
	},
	Headers: func(_ string, vals engine.Values) map[string]string {
		v := vals["version"]
		if !v.IsKnown() || v.IsNull() {
			return nil
		}
		var n big.Float
		if err := v.As(&n); err != nil {
			return nil
		}
		i, _ := n.Int64()
		return map[string]string{"If-Match": `"` + big.NewInt(i).String() + `"`}
	},
}
