// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

func TestInexplicitNumberOrderCount(t *testing.T) {
	in := engine.Object{"ordering_groups": []any{map[string]any{"count_requested": json.Number("3")}, "x"}}
	out, err := inexplicitNumberOrderHooks.AfterRead(context.Background(), in)
	g := out["ordering_groups"].([]any)                       //nolint:forcetypeassert
	first := g[0].(map[string]any)                            //nolint:forcetypeassert
	orig := in["ordering_groups"].([]any)[0].(map[string]any) //nolint:forcetypeassert
	if err != nil || first["count_requested"] != "3" || g[1] != "x" || orig["count_requested"] != json.Number("3") {
		t.Errorf("out %v, err %v", out, err)
	}
}
