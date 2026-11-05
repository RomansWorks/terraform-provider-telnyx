// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"strings"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// Inboxes take a username on create but answer only with the full address.
// The read hook puts the address's local part back as username, so a
// generated username is known and import does not plan a replacement.
var emailInboxHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		addr, _ := obj["address"].(string)
		at := strings.LastIndex(addr, "@")
		if at <= 0 {
			return obj, nil
		}
		return withField(obj, "username", addr[:at]), nil
	},
}

// withField returns a copy of obj with one more field.
func withField(obj engine.Object, name string, v any) engine.Object {
	out := make(engine.Object, len(obj)+1)
	for k, val := range obj {
		out[k] = val
	}
	out[name] = v
	return out
}
