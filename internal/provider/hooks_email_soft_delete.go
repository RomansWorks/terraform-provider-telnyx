// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// goneWithStatus returns a read hook for objects the API soft-deletes: they
// are still returned, with this status, after they were deleted. Such an
// object is treated as deleted outside Terraform: the hook answers with a
// not-found error, which removes it from state on read.
func goneWithStatus(status string) func(context.Context, engine.Object) (engine.Object, error) {
	return func(_ context.Context, obj engine.Object) (engine.Object, error) {
		if s, _ := obj["status"].(string); s == status {
			return nil, &client.Error{Method: http.MethodGet, StatusCode: http.StatusNotFound, Body: "the object is " + status}
		}
		return obj, nil
	}
}

// Suppressions deleted in the portal or the API stay readable as removed.
var emailSuppressionHooks = engine.Hooks{AfterRead: goneWithStatus("removed")}

// Released dedicated IPs stay readable as released until their quarantine
// ends.
var emailDedicatedIPHooks = engine.Hooks{AfterRead: goneWithStatus("released")}
