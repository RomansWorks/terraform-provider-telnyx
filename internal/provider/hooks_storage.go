// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// KV keys are written with the raw value as the request body and an
// optional ttl_secs query parameter; both create and update are that PUT.
var storageKVKeyHooks = engine.Hooks{
	Send: func(ctx context.Context, c *client.Client, r *engine.Resource, op string, plan, pathVals engine.Values) (engine.Object, error) {
		var value string
		if err := plan["value"].As(&value); err != nil {
			return nil, err
		}
		p, err := r.Path(r.Create.Path, pathVals)
		if err != nil {
			return nil, err
		}
		var query url.Values
		if ttl := plan["ttl_secs"]; ttl.IsKnown() && !ttl.IsNull() {
			var n big.Float
			if err := ttl.As(&n); err != nil {
				return nil, err
			}
			i, _ := n.Int64()
			query = url.Values{"ttl_secs": {fmt.Sprint(i)}}
		}
		if _, err := c.Send(ctx, http.MethodPut, p, query, "application/octet-stream", []byte(value)); err != nil {
			return nil, err
		}
		// The response is empty; the read after the write fills the state.
		return engine.Object{}, nil
	},
}

// A bucket's SSL certificate is uploaded as two files of a multipart form.
// The PEM text comes from configuration; the API returns only the
// certificate's details, so the configured PEM is kept in state.
var storageBucketSSLCertificateHooks = engine.Hooks{
	Send: func(ctx context.Context, c *client.Client, r *engine.Resource, op string, plan, pathVals engine.Values) (engine.Object, error) {
		var cert, key string
		if err := plan["certificate_pem"].As(&cert); err != nil {
			return nil, err
		}
		if err := plan["private_key_pem"].As(&key); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		for _, f := range []struct{ field, name, content string }{
			{"certificate", "certificate.pem", cert},
			{"private_key", "private_key.pem", key},
		} {
			h := textproto.MIMEHeader{}
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, f.field, f.name))
			h.Set("Content-Type", "application/x-pem-file")
			part, err := w.CreatePart(h)
			if err != nil {
				return nil, err
			}
			if _, err := part.Write([]byte(f.content)); err != nil {
				return nil, err
			}
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		p, err := r.Path(r.Create.Path, pathVals)
		if err != nil {
			return nil, err
		}
		out, err := c.Send(ctx, http.MethodPut, p, nil, w.FormDataContentType(), buf.Bytes())
		if err != nil {
			return nil, err
		}
		return engine.Decode(out, r.Create.Envelope)
	},
}

// storageProvisionPoll is how often a provisioning KV namespace or SQL
// database is checked.
var storageProvisionPoll = 2 * time.Second

// KV namespaces and SQL databases are provisioned asynchronously, and their
// keys and queries answer 409 until status is provision_ok, so create waits.
var storageProvisionHooks = engine.Hooks{
	Wait: func(ctx context.Context, c *client.Client, r *engine.Resource, obj engine.Object) (engine.Object, error) {
		for {
			switch status, _ := obj["status"].(string); status {
			case "provision_ok", "":
				return obj, nil
			case "provision_failed", "deleting", "delete_failed":
				return nil, fmt.Errorf("%s %v is %s; destroy it and create it again", r.TypeName, obj["id"], status)
			}
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%s %v is still %v: %w", r.TypeName, obj["id"], obj["status"], ctx.Err())
			case <-time.After(storageProvisionPoll):
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
