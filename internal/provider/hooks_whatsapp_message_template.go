// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// WhatsApp message templates are created with waba_id but read back with
// the account as whatsapp_business_account.id. The read hook copies it to
// waba_id, so an imported template keeps its account and is not replaced.
var whatsappMessageTemplateHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		if _, ok := obj["waba_id"]; ok {
			return obj, nil
		}
		acct, _ := obj["whatsapp_business_account"].(map[string]any)
		id, ok := acct["id"]
		if !ok {
			return obj, nil
		}
		out := make(engine.Object, len(obj)+1)
		for k, v := range obj {
			out[k] = v
		}
		out["waba_id"] = id
		return out, nil
	},
}
