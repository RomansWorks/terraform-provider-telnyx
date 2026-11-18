// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// recorder is a fake API that answers every request with one body and
// records what it was sent.
type recorder struct {
	mu   sync.Mutex
	reqs []string
}

func (r *recorder) log() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.reqs...)
}

func newRecorder(t *testing.T, answer string) (*recorder, *client.Client) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		rec.mu.Lock()
		rec.reqs = append(rec.reqs, fmt.Sprintf("%s %s %s", req.Method, req.URL.Path, body))
		rec.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	return rec, c
}

func boxToken() *Ephemeral {
	return &Ephemeral{
		TypeName: "telnyx_box_token",
		Open:     &Op{Method: "POST", Path: "/boxes/{box_id}/token", Envelope: true},
		Attrs: []*Attr{
			{Name: "box_id", Kind: KindString, Required: true, Local: true, PathParams: []string{"box_id"}},
			{Name: "token", JSON: "token", Kind: KindString, Computed: true, InResponse: true},
			{Name: "ttl_secs", JSON: "ttl_secs", Kind: KindInt, Optional: true, InCreate: true},
		},
	}
}

// TestEphemeralArguments sends the request fields the configuration sets,
// and only those, and keeps them as configured.
func TestEphemeralArguments(t *testing.T) {
	rec, c := newRecorder(t, `{"data":{"token":"tok"}}`)
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV6ProviderFactories: providers(c, &testProvider{ephemerals: []*Ephemeral{boxToken()}}),
		Steps: []rtest.TestStep{
			ephemeralStep(rtest.TestStep{
				Config: `
ephemeral "telnyx_box_token" "a" {
  box_id   = "b1"
  ttl_secs = 60
}
ephemeral "telnyx_box_token" "b" {
  box_id = "b2"
}
provider "echo" {
  data = "${ephemeral.telnyx_box_token.a.token}/${ephemeral.telnyx_box_token.a.ttl_secs}"
}
resource "echo" "t" {}`,
				Check: func(s *terraform.State) error {
					if err := rtest.TestCheckResourceAttr("echo.t", "data", "tok/60")(s); err != nil {
						return err
					}
					seen := map[string]bool{}
					for _, r := range rec.log() {
						seen[r] = true
					}
					if !seen[`POST /boxes/b1/token {"ttl_secs":60}`] || !seen[`POST /boxes/b2/token {}`] {
						return fmt.Errorf("requests %v", rec.log())
					}
					return nil
				},
			}),
		},
	})
}

func TestEphemeralArgumentsSchema(t *testing.T) {
	s := boxToken().Schema()
	ttl := s.Attributes["ttl_secs"]
	if !ttl.IsOptional() || ttl.IsComputed() || ttl.IsSensitive() {
		t.Errorf("ttl_secs: optional %v computed %v sensitive %v", ttl.IsOptional(), ttl.IsComputed(), ttl.IsSensitive())
	}
	if tok := s.Attributes["token"]; !tok.IsComputed() || !tok.IsSensitive() {
		t.Errorf("token: computed %v sensitive %v", tok.IsComputed(), tok.IsSensitive())
	}
	if id := s.Attributes["box_id"]; !id.IsRequired() {
		t.Error("box_id is not required")
	}
}

// keySecret can be created or replaced but never read back or deleted.
func keySecret() *Resource {
	return &Resource{
		TypeName:   "telnyx_key_secret",
		Create:     &Op{Method: "POST", Path: "/key_secrets", Envelope: true},
		IDAttr:     "name",
		Import:     []string{"name"},
		Unreadable: true,
		Attrs: []*Attr{
			{Name: "name", JSON: "name", Kind: KindString, Required: true, ForceNew: true, InCreate: true, InResponse: true},
			{Name: "value_wo", JSON: "value", Kind: KindString, Required: true, Sensitive: true, InCreate: true, WriteOnly: true, Version: "value_wo_version"},
			{Name: "value_wo_version", Kind: KindInt, Optional: true, ForceNew: true, Local: true},
		},
	}
}

// TestUnreadable creates with one request, refreshes and destroys without
// any, replaces when the version changes, and refuses import.
func TestUnreadable(t *testing.T) {
	rec, c := newRecorder(t, `{"data":{"name":"a","value":"REDACTED"}}`)
	cfg := func(version int) string {
		return fmt.Sprintf(`resource "telnyx_key_secret" "s" {
  name             = "a"
  value_wo         = "v%d"
  value_wo_version = %d
}`, version, version)
	}
	posts := func(want int) rtest.TestCheckFunc {
		return func(*terraform.State) error {
			if got := len(rec.log()); got != want {
				return fmt.Errorf("want %d requests, got %v", want, rec.log())
			}
			return nil
		}
	}
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		ProtoV6ProviderFactories: factories(c, keySecret()),
		CheckDestroy: func(*terraform.State) error {
			if got := len(rec.log()); got != 2 {
				return fmt.Errorf("destroy sent a request: %v", rec.log())
			}
			return nil
		},
		Steps: []rtest.TestStep{
			{Config: cfg(1), Check: rtest.ComposeTestCheckFunc(posts(1),
				rtest.TestCheckResourceAttr("telnyx_key_secret.s", "name", "a"),
				rtest.TestCheckNoResourceAttr("telnyx_key_secret.s", "value_wo"))},
			{Config: cfg(1), PlanOnly: true},
			{Config: cfg(2), Check: func(s *terraform.State) error {
				log := rec.log()
				var body map[string]any
				if len(log) != 2 || json.Unmarshal([]byte(log[1][len("POST /key_secrets "):]), &body) != nil || body["value"] != "v2" {
					return fmt.Errorf("requests %v", log)
				}
				return nil
			}},
			{
				ResourceName:  "telnyx_key_secret.s",
				ImportState:   true,
				ImportStateId: "a",
				ExpectError:   regexpMust(`Import not supported`),
			},
		},
	})
}
