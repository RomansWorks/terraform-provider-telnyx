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

// agentGatewayPoll is how often a starting Agent Gateway is checked.
var agentGatewayPoll = 10 * time.Second

// An Agent Gateway starts asynchronously, and its sub-objects answer 409
// until it is ready, so create waits for ready.
var agentGatewayHooks = engine.Hooks{
	Wait: func(ctx context.Context, c *client.Client, r *engine.Resource, obj engine.Object) (engine.Object, error) {
		for {
			switch status, _ := obj["status"].(string); status {
			case "ready":
				return obj, nil
			case "failed", "deleting":
				return nil, fmt.Errorf("agent gateway %v is %s; destroy it and create it again", obj["id"], status)
			}
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("agent gateway %v is still %v: %w", obj["id"], obj["status"], ctx.Err())
			case <-time.After(agentGatewayPoll):
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

// Tool permissions are a grant the API only replaces: destroy grants none.
var agentGatewayToolPermissionsHooks = engine.Hooks{
	Delete: func(ctx context.Context, c *client.Client, r *engine.Resource, state engine.Values) error {
		p, err := r.Path(r.Update.Path, state)
		if err != nil {
			return err
		}
		_, err = c.Send(ctx, r.Update.Method, p, nil, "application/json", []byte(`{"servers":[]}`))
		return err
	},
}
