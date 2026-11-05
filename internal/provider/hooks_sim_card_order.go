// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// A SIM card order is created with address_id but answers with the address
// in order_address. The read hook copies order_address.id to address_id, so
// an imported order matches its configuration instead of being replaced by
// a new, billable order.
var simCardOrderHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		addr, ok := obj["order_address"].(map[string]any)
		if !ok || addr["id"] == nil {
			return obj, nil
		}
		out := make(engine.Object, len(obj)+1)
		for k, v := range obj {
			out[k] = v
		}
		out["address_id"] = addr["id"]
		return out, nil
	},
}
