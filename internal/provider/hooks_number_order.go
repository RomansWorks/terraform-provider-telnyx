// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// numberOrderPoll is how often a pending number order is checked.
var numberOrderPoll = 5 * time.Second

// A number order completes asynchronously: wait until it succeeds. The API
// cannot cancel an order, so destroy names the numbers that stay.
var numberOrderHooks = engine.Hooks{
	DeleteWarning: func(state engine.Values) string {
		var numbers []string
		var list []tftypes.Value
		if v := state["phone_numbers"]; v.IsKnown() && !v.IsNull() && v.As(&list) == nil {
			for _, item := range list {
				var f map[string]tftypes.Value
				var n string
				if item.As(&f) == nil && f["phone_number"].As(&n) == nil && n != "" {
					numbers = append(numbers, n)
				}
			}
		}
		msg := "The Telnyx API cannot cancel a number order, so it was only removed from Terraform state."
		if len(numbers) > 0 {
			msg += " These numbers stay in your account until you release them: " + strings.Join(numbers, ", ") + "."
		}
		return msg
	},
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
