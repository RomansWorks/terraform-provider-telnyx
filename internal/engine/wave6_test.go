// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// limitResource is identified by product and period, as Telnyx spend limits
// are.
func limitResource() *Resource {
	return &Resource{
		TypeName: "telnyx_limit",
		IDAttr:   "product",
		ReadFrom: &ReadFrom{Op: &Op{Method: "GET", Path: "/limits", Envelope: true}, Key: "product", Match: []string{"period"}},
		Attrs: []*Attr{
			{Name: "product", JSON: "product", Kind: KindString, Required: true, PathParams: []string{"product"}},
			{Name: "period", JSON: "period", Kind: KindString, Required: true},
			{Name: "amount", JSON: "amount", Kind: KindString, Computed: true, InResponse: true},
		},
	}
}

// TestReadFromMatch finds the element whose key and match fields both
// equal the state.
func TestReadFromMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"product":"voice","period":"daily","amount":"1"},{"product":"voice","period":"monthly","amount":"20"}]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	d := limitResource()
	obj, err := Get(context.Background(), c, d, Values{"product": str("voice"), "period": str("monthly")})
	if err != nil || obj["amount"] != "20" {
		t.Errorf("monthly: %v %v", obj, err)
	}
	_, err = Get(context.Background(), c, d, Values{"product": str("voice"), "period": str("weekly")})
	if !client.IsNotFound(err) {
		t.Errorf("weekly: want not found, got %v", err)
	}
	_, err = Get(context.Background(), c, d, Values{"product": str("voice"), "period": tftypes.NewValue(tftypes.String, nil)})
	if err == nil {
		t.Error("unknown period: want an error")
	}
}

// TestOpQuery sends only the set attributes the operation names.
func TestOpQuery(t *testing.T) {
	attrs := limitResource().Attrs
	op := &Op{Method: "DELETE", Path: "/limits/{product}", Query: []string{"period"}}
	q := opQuery(attrs, op, Values{"product": str("voice"), "period": str("monthly"), "amount": str("2")})
	if q.Encode() != "period=monthly" {
		t.Errorf("query %s", q.Encode())
	}
	if q := opQuery(attrs, op, Values{"period": tftypes.NewValue(tftypes.String, nil)}); q != nil {
		t.Errorf("null period: %s", q.Encode())
	}
	if q := opQuery(attrs, &Op{Method: "DELETE", Path: "/limits/{product}"}, Values{"period": str("monthly")}); q != nil {
		t.Errorf("no query: %s", q.Encode())
	}
}

// TestOnceKeepsCreateValue keeps a value only create returns, except in a
// data source, which shows what the API says.
func TestOnceKeepsCreateValue(t *testing.T) {
	attrs := []*Attr{{Name: "secret", JSON: "secret", Kind: KindString, Computed: true, InResponse: true, Once: true}}
	prior := map[string]tftypes.Value{"secret": str("s3cret")}
	for _, m := range []mode{modeApply, modeRead} {
		v, err := objectFromJSON(attrs, Object{"secret": nil}, prior, m)
		if err != nil || !fields(v)["secret"].Equal(str("s3cret")) {
			t.Errorf("mode %d: %v %v", m, v, err)
		}
	}
	v, err := objectFromJSON(attrs, Object{"secret": nil}, prior, modeData)
	if err != nil || !fields(v)["secret"].IsNull() {
		t.Errorf("data: %v %v", v, err)
	}
	// Without a value in state, the API's value is used.
	v, err = objectFromJSON(attrs, Object{"secret": "new"}, nil, modeRead)
	if err != nil || !fields(v)["secret"].Equal(str("new")) {
		t.Errorf("no prior: %v %v", v, err)
	}
}

// TestDeleteHookAndGoneRead: a Delete hook replaces the generic delete, its
// not-found error is ignored, and a not-found error from a read hook removes
// the object from state.
func TestDeleteHookAndGoneRead(t *testing.T) {
	api, c := newCollectionAPI(t)
	d := secret()
	var hookDeletes int
	gone := false
	d.Hooks.Delete = func(_ context.Context, _ *client.Client, _ *Resource, state Values) error {
		hookDeletes++
		return &client.Error{Method: "DELETE", Path: "/secrets", StatusCode: http.StatusNotFound}
	}
	d.Hooks.AfterRead = func(_ context.Context, obj Object) (Object, error) {
		if gone {
			return nil, &client.Error{Method: "GET", Path: "/secrets", StatusCode: http.StatusNotFound}
		}
		return obj, nil
	}
	cfg := `resource "telnyx_secret" "s" {
  name  = "a"
  value = "v"
}`
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, d),
		CheckDestroy: func(*terraform.State) error {
			if hookDeletes != 1 || api.sent("DELETE") != 0 {
				return fmt.Errorf("hook deletes %d, requests %v", hookDeletes, api.log())
			}
			return nil
		},
		Steps: []rtest.TestStep{
			{Config: cfg},
			{
				// The read hook says the object is gone: plan creates it again.
				PreConfig:          func() { gone = true },
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{PreConfig: func() { gone = false }, Config: cfg},
		},
	})
}

// TestImportDoesNotReplaceForUnreturnedValues: after import, a ForceNew
// value the API does not return is null; setting it in configuration
// updates the state instead of replacing the object.
func TestImportDoesNotReplaceForUnreturnedValues(t *testing.T) {
	_, c := newCollectionAPI(t)
	cfg := `resource "telnyx_secret" "s" {
  name  = "a"
  value = "v"
}`
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: factories(c, secret()),
		Steps: []rtest.TestStep{
			{Config: cfg},
			{
				// Forget it, keeping it in the API, then import it.
				Config: `removed {
  from = telnyx_secret.s
  lifecycle {
    destroy = false
  }
}`,
			},
			{
				Config:             cfg,
				ResourceName:       "telnyx_secret.s",
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateId:      "s1",
			},
			{
				Config: cfg,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_secret.s", plancheck.ResourceActionUpdate)},
				},
			},
			{
				// Once set, a change replaces as before.
				Config: `resource "telnyx_secret" "s" {
  name  = "a"
  value = "w"
}`,
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_secret.s", plancheck.ResourceActionReplace)},
				},
			},
		},
	})
}
