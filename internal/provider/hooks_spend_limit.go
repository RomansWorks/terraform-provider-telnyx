// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// A spend limit update replaces the limit: the API needs exactly one of
// amount and unlimited in every request, also when only the reason
// changes, and refuses a null for the one not used.
//
// The list also has an entry for every product and period without a limit
// (limit null): such an entry means the limit is gone.
var spendLimitHooks = engine.Hooks{
	AfterRead: func(_ context.Context, obj engine.Object) (engine.Object, error) {
		if limit, ok := obj["limit"]; ok && limit == nil {
			return nil, &client.Error{Method: http.MethodGet, Path: "/spend_limits", StatusCode: http.StatusNotFound,
				Body: "no limit is set for this product and period"}
		}
		return obj, nil
	},
	BeforeRequest: func(_ context.Context, op string, body engine.Object, plan engine.Values) error {
		delete(body, "amount")
		delete(body, "unlimited")
		var unlimited bool
		if v := plan["unlimited"]; v.IsKnown() && !v.IsNull() && v.As(&unlimited) == nil && unlimited {
			body["unlimited"] = true
			return nil
		}
		var amount big.Float
		if v := plan["amount"]; v.IsKnown() && !v.IsNull() && v.As(&amount) == nil {
			body["amount"] = json.Number(amount.Text('g', -1))
		}
		return nil
	},
}
