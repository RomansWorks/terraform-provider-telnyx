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
