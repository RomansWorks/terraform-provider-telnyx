// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// TestEachCursorPaging follows both cursor shapes Telnyx uses until the
// cursor is empty.
func TestEachCursorPaging(t *testing.T) {
	pages := map[string]string{
		"":   `{"data":[{"id":"a"},{"id":"b"}],"meta":{"cursors":{"after":"c1"}}}`,
		"c1": `{"data":[{"id":"c"},{"id":"d"}],"meta":{"end_cursor":"c2"}}`,
		"c2": `{"data":[{"id":"e"}],"meta":{"cursors":{"after":""}}}`,
	}
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queries = append(queries, q.Encode())
		if q.Get("page[size]") != "2" {
			t.Errorf("page size: %s", q.Encode())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pages[q.Get("page[after]")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	op := &Op{Method: "GET", Path: "/things", Envelope: true, Paging: PagingCursor, PageSize: 2}
	var ids []string
	err = Each(context.Background(), c, op, "/things", nil, func(o Object) bool {
		ids = append(ids, o["id"].(string)) //nolint:forcetypeassert
		return true
	})
	if err != nil || strings.Join(ids, ",") != "a,b,c,d,e" || len(queries) != 3 {
		t.Errorf("ids %v, err %v, queries %v", ids, err, queries)
	}
	if strings.Contains(queries[0], "after") {
		t.Errorf("first page sent a cursor: %s", queries[0])
	}
}

func TestNextCursor(t *testing.T) {
	for in, want := range map[string]string{
		`{"meta":{"cursors":{"after":"x"}}}`: "x",
		`{"meta":{"end_cursor":"y"}}`:        "y",
		`{"meta":{"end_cursor":null}}`:       "",
		`{"meta":{"page_number":1}}`:         "",
		`not json`:                           "",
	} {
		if got := nextCursor([]byte(in)); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestWithFixedID(t *testing.T) {
	r := &Resource{FixedID: "settings"}
	in := Object{"channels": json.Number("3")}
	out := r.withFixedID(in)
	if out["id"] != "settings" || out["channels"] != json.Number("3") || in["id"] != nil {
		t.Errorf("out %v, in %v", out, in)
	}
	if got := (&Resource{}).withFixedID(in); got["id"] != nil {
		t.Errorf("no fixed id: %v", got)
	}
}

// textAPI answers a plain-text token and a lookup by query parameters.
func textAPI(t *testing.T) (*client.Client, *[]string) {
	t.Helper()
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Path {
		case "/creds/c1/token":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("jwt-123\n"))
		case "/status":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"connection_id":"` + r.URL.Query().Get("connection_id") + `","registered":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	return c, &seen
}

func TestEphemeralText(t *testing.T) {
	c, _ := textAPI(t)
	token := &Ephemeral{
		TypeName: "telnyx_cred_token",
		Open:     &Op{Method: "POST", Path: "/creds/{id}/token", Text: true},
		Attrs: []*Attr{
			{Name: "id", Kind: KindString, Required: true, Local: true, PathParams: []string{"id"}},
			{Name: "token", JSON: "value", Kind: KindString, Computed: true, InResponse: true},
		},
	}
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV6ProviderFactories: providers(c, &testProvider{ephemerals: []*Ephemeral{token}}),
		Steps: []rtest.TestStep{ephemeralStep(rtest.TestStep{
			Config: `
ephemeral "telnyx_cred_token" "t" {
  id = "c1"
}
provider "echo" {
  data = ephemeral.telnyx_cred_token.t.token
}
resource "echo" "t" {}`,
			Check: rtest.TestCheckResourceAttr("echo.t", "data", "jwt-123"),
		})},
	})
}

func TestDataSourceQuery(t *testing.T) {
	c, seen := textAPI(t)
	status := &Resource{
		TypeName:   "telnyx_status",
		Read:       &Op{Method: "GET", Path: "/status"},
		DataSource: true,
		NoResource: true,
		Import:     []string{},
		Attrs: []*Attr{
			{Name: "connection_id", JSON: "connection_id", Kind: KindString, Computed: true, InResponse: true, Query: "connection_id"},
			{Name: "credential_type", Kind: KindString, Required: true, Local: true, Query: "credential_type"},
			{Name: "username", Kind: KindString, Local: true, Query: "username"},
			{Name: "registered", JSON: "registered", Kind: KindBool, Computed: true, InResponse: true},
		},
	}
	rtest.UnitTest(t, rtest.TestCase{
		ProtoV6ProviderFactories: providers(c, &testProvider{res: []*Resource{status}}),
		Steps: []rtest.TestStep{{
			Config: `
data "telnyx_status" "s" {
  connection_id   = "conn-1"
  credential_type = "telephony_credential"
}`,
			Check: rtest.ComposeAggregateTestCheckFunc(
				rtest.TestCheckResourceAttr("data.telnyx_status.s", "connection_id", "conn-1"),
				rtest.TestCheckResourceAttr("data.telnyx_status.s", "credential_type", "telephony_credential"),
				rtest.TestCheckResourceAttr("data.telnyx_status.s", "registered", "true"),
				rtest.TestCheckNoResourceAttr("data.telnyx_status.s", "username"),
			),
		}},
	})
	for _, s := range *seen {
		if s != "GET /status?connection_id=conn-1&credential_type=telephony_credential" {
			t.Errorf("request: %s", s)
		}
	}
}

func TestMasked(t *testing.T) {
	attrs := []*Attr{{Name: "password", JSON: "password", Kind: KindString, Required: true, Masked: true, InResponse: true}}
	prior := map[string]tftypes.Value{"password": tftypes.NewValue(tftypes.String, "s3cret")}
	for raw, want := range map[string]string{"********": "s3cret", "changed": "changed"} {
		v, err := objectFromJSON(attrs, Object{"password": raw}, prior, modeRead)
		if err != nil {
			t.Fatal(err)
		}
		var got string
		_ = fields(v)["password"].As(&got)
		if got != want {
			t.Errorf("%s: got %q, want %q", raw, got, want)
		}
	}
	if isMask("") || isMask("a*") || !isMask("****") {
		t.Error("isMask")
	}
}

func TestFormBody(t *testing.T) {
	got, err := formBody(Object{"MaxSize": json.Number("5"), "FriendlyName": "q", "Skip": nil, "On": true})
	if err != nil || string(got) != "FriendlyName=q&MaxSize=5&On=true" {
		t.Errorf("got %s, %v", got, err)
	}
	if _, err := formBody(Object{"nested": map[string]any{}}); err == nil {
		t.Error("nested field: want an error")
	}
}

func TestDecodeOnly(t *testing.T) {
	obj, err := decodeOnly([]byte(`{"data":[{"id":"po1"}]}`))
	if err != nil || obj["id"] != "po1" {
		t.Errorf("one: %v %v", obj, err)
	}
	if _, err := decodeOnly([]byte(`{"data":[{"id":"a"},{"id":"b"}]}`)); err == nil || !strings.Contains(err.Error(), "a, b") {
		t.Errorf("two: %v", err)
	}
}
