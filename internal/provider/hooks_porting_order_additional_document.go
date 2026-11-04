// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// The API adds additional documents in batches: it takes a list and returns
// a list. Each Terraform resource is one document, sent as a list of one.
var portingOrderAdditionalDocumentHooks = engine.Hooks{
	Create: func(ctx context.Context, c *client.Client, r *engine.Resource, plan engine.Values) (engine.Object, error) {
		var id, typ string
		if err := plan["document_id"].As(&id); err != nil {
			return nil, err
		}
		if err := plan["document_type"].As(&typ); err != nil {
			return nil, err
		}
		p, err := r.Path(r.Create.Path, plan)
		if err != nil {
			return nil, err
		}
		body, err := json.Marshal(map[string]any{
			"additional_documents": []map[string]string{{"document_id": id, "document_type": typ}},
		})
		if err != nil {
			return nil, err
		}
		out, err := c.Send(ctx, http.MethodPost, p, nil, "application/json", body)
		if err != nil {
			return nil, err
		}
		return firstOfList(out)
	},
}

// firstOfList returns the only object of a {"data": [...]} response.
func firstOfList(out []byte) (engine.Object, error) {
	dec := json.NewDecoder(bytes.NewReader(out))
	dec.UseNumber()
	var body struct {
		Data []engine.Object `json:"data"`
	}
	if err := dec.Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if len(body.Data) != 1 {
		return nil, fmt.Errorf("expected one object in the response, got %d", len(body.Data))
	}
	return body.Data[0], nil
}
