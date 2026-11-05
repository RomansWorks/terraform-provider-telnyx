// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

func TestEmailInboxUsername(t *testing.T) {
	for addr, want := range map[string]any{
		"support@mail.example.com": "support",
		"a.b@x@y":                  "a.b@x",
		"":                         nil,
		"@example.com":             nil,
	} {
		obj, err := emailInboxHooks.AfterRead(context.Background(), engine.Object{"address": addr})
		if err != nil {
			t.Fatal(err)
		}
		if got := obj["username"]; got != want {
			t.Errorf("%q: username %v, want %v", addr, got, want)
		}
	}
}

func TestEmailDomainInboundEnabled(t *testing.T) {
	obj, err := emailDomainHooks.AfterRead(context.Background(), engine.Object{"inbound": map[string]any{"enabled": true}})
	if err != nil || obj["inbound_enabled"] != true {
		t.Errorf("inbound_enabled: %v %v", obj["inbound_enabled"], err)
	}
	obj, _ = emailDomainHooks.AfterRead(context.Background(), engine.Object{"id": "d1"})
	if _, ok := obj["inbound_enabled"]; ok {
		t.Errorf("inbound_enabled without inbound: %v", obj)
	}
}
