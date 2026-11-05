// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// The API adds and removes DIR phone numbers in batches, named in the
// request body. Each Terraform resource is one number, sent as a list of
// one.
var displayIdentityRecordPhoneNumberHooks = engine.Hooks{
	BeforeRequest: phoneNumberAsList,
	Delete: func(ctx context.Context, c *client.Client, r *engine.Resource, state engine.Values) error {
		var number string
		if err := state["phone_number"].As(&number); err != nil {
			return err
		}
		p, err := r.Path(r.Delete.Path, state)
		if err != nil {
			return err
		}
		body, err := json.Marshal(map[string]any{"phone_numbers": []string{number}})
		if err != nil {
			return err
		}
		_, err = c.Send(ctx, http.MethodDelete, p, nil, "application/json", body)
		return err
	},
}

// phoneNumberAsList sends the planned phone_number as phone_numbers, a list
// of one, on create.
func phoneNumberAsList(_ context.Context, op string, body engine.Object, plan engine.Values) error {
	if op != "create" {
		return nil
	}
	var number string
	if err := plan["phone_number"].As(&number); err != nil {
		return err
	}
	body["phone_numbers"] = []string{number}
	return nil
}
