// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// numberOrderPoll is how often a pending number order is checked.
var numberOrderPoll = 5 * time.Second

// A number order completes asynchronously: wait until it succeeds.
var numberOrderHooks = engine.Hooks{
	Wait: func(ctx context.Context, c *client.Client, r *engine.Resource, obj engine.Object) (engine.Object, error) {
		for {
			switch status, _ := obj["status"].(string); status {
			case "success":
				return obj, nil
			case "failure":
				return nil, fmt.Errorf("number order %v failed; check its phone numbers and regulatory requirements in the Telnyx portal", obj["id"])
			}
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("number order %v is still %v: %w", obj["id"], obj["status"], ctx.Err())
			case <-time.After(numberOrderPoll):
			}
			id, _ := obj["id"].(string)
			next, err := engine.Get(ctx, c, r, engine.Values{"id": tftypes.NewValue(tftypes.String, id)})
			if err != nil {
				return nil, err
			}
			obj = next
		}
	},
}
