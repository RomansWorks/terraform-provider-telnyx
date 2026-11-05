// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

func TestSimCardOrderAddressID(t *testing.T) {
	ctx := context.Background()
	got, err := simCardOrderHooks.AfterRead(ctx, engine.Object{"id": "o1", "order_address": map[string]any{"id": "a1"}})
	if err != nil || got["address_id"] != "a1" {
		t.Errorf("got %v, %v", got, err)
	}
	got, err = simCardOrderHooks.AfterRead(ctx, engine.Object{"id": "o1"})
	if err != nil || got["address_id"] != nil {
		t.Errorf("without an address: %v, %v", got, err)
	}
}
