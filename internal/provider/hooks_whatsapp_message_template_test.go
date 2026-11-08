// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

func TestWhatsappMessageTemplateWabaID(t *testing.T) {
	in := engine.Object{"id": "t1", "whatsapp_business_account": map[string]any{"id": "w1"}}
	out, err := whatsappMessageTemplateHooks.AfterRead(context.Background(), in)
	if err != nil || out["waba_id"] != "w1" || in["waba_id"] != nil {
		t.Errorf("out %v, in %v, err %v", out, in, err)
	}
	same, _ := whatsappMessageTemplateHooks.AfterRead(context.Background(), engine.Object{"waba_id": "w2"})
	if same["waba_id"] != "w2" {
		t.Errorf("kept: %v", same)
	}
}
