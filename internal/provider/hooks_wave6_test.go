// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/provider/generated"
)

func tfString(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

func TestDisplayIdentityRecordCallReasons(t *testing.T) {
	in := engine.Object{"id": "d1", "call_reasons": []any{
		map[string]any{"reason": "Appointment reminders", "created_at": "2026-01-01"},
		"Delivery updates",
	}}
	out, err := displayIdentityRecordHooks.AfterRead(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(out["call_reasons"])
	if string(got) != `["Appointment reminders","Delivery updates"]` || out["id"] != "d1" {
		t.Errorf("got %s", got)
	}
	if _, ok := in["call_reasons"].([]any)[0].(map[string]any); !ok {
		t.Error("the hook changed its input")
	}
}

func TestPhoneNumberAsList(t *testing.T) {
	body := engine.Object{}
	plan := engine.Values{"phone_number": tfString("+13125550100")}
	if err := phoneNumberAsList(context.Background(), "create", body, plan); err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(body)
	if string(got) != `{"phone_numbers":["+13125550100"]}` {
		t.Errorf("create body %s", got)
	}
	body = engine.Object{}
	if err := phoneNumberAsList(context.Background(), "update", body, plan); err != nil || len(body) != 0 {
		t.Errorf("update body %v %v", body, err)
	}
}

func TestSpendLimitBody(t *testing.T) {
	null := func(typ tftypes.Type) tftypes.Value { return tftypes.NewValue(typ, nil) }
	for name, tc := range map[string]struct {
		plan engine.Values
		want string
	}{
		"amount":                  {engine.Values{"amount": tftypes.NewValue(tftypes.Number, big.NewFloat(25.5)), "unlimited": null(tftypes.Bool)}, `{"amount":25.5,"reason":"r"}`},
		"unlimited":               {engine.Values{"amount": null(tftypes.Number), "unlimited": tftypes.NewValue(tftypes.Bool, true)}, `{"reason":"r","unlimited":true}`},
		"amount, unlimited false": {engine.Values{"amount": tftypes.NewValue(tftypes.Number, big.NewFloat(3)), "unlimited": tftypes.NewValue(tftypes.Bool, false)}, `{"amount":3,"reason":"r"}`},
	} {
		// Only the reason changed: the request still carries the limit, and
		// never a null for the field not used.
		body := engine.Object{"reason": "r", "amount": nil}
		if err := spendLimitHooks.BeforeRequest(context.Background(), "update", body, tc.plan); err != nil {
			t.Fatal(err)
		}
		got, _ := json.Marshal(body)
		if string(got) != tc.want {
			t.Errorf("%s: got %s, want %s", name, got, tc.want)
		}
	}
}

func TestDisplayIdentityRecordPhoneNumberDelete(t *testing.T) {
	var method, path, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		method, path, body = r.Method, r.URL.Path, string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":["+13125550100"],"meta":{"errors":[]}}`))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	d := generated.DisplayIdentityRecordPhoneNumber()
	state := engine.Values{"dir_id": tfString("dir1"), "id": tfString("n1"), "phone_number": tfString("+13125550100")}
	if err := displayIdentityRecordPhoneNumberHooks.Delete(context.Background(), c, d, state); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodDelete || path != "/dir/dir1/phone_numbers" || body != `{"phone_numbers":["+13125550100"]}` {
		t.Errorf("%s %s %s", method, path, body)
	}
}

func TestSpendLimitGone(t *testing.T) {
	_, err := spendLimitHooks.AfterRead(context.Background(), engine.Object{"product": "voice", "limit": nil})
	if !client.IsNotFound(err) {
		t.Errorf("limit null: want not found, got %v", err)
	}
	obj := engine.Object{"product": "voice", "limit": map[string]any{"amount": "5"}}
	if out, err := spendLimitHooks.AfterRead(context.Background(), obj); err != nil || out["product"] != "voice" {
		t.Errorf("limit set: %v %v", out, err)
	}
}

func TestSpendLimitReadsLimit(t *testing.T) {
	out, err := spendLimitHooks.AfterRead(context.Background(), engine.Object{"limit": map[string]any{"amount": "50.00", "unlimited": false}})
	if err != nil || out["amount"] != "50.00" || out["unlimited"] != nil {
		t.Errorf("amount: %v %v", out, err)
	}
	out, err = spendLimitHooks.AfterRead(context.Background(), engine.Object{"limit": map[string]any{"amount": nil, "unlimited": true}})
	if err != nil || out["unlimited"] != true || out["amount"] != nil {
		t.Errorf("unlimited: %v %v", out, err)
	}
}
