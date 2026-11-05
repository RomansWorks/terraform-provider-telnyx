// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// An Agent Gateway is created pending; create waits until it is ready.
func TestAgentGatewayWaitsForReady(t *testing.T) {
	defer func(d time.Duration) { agentGatewayPoll = d }(agentGatewayPoll)
	agentGatewayPoll = 10 * time.Millisecond

	var gw map[string]any
	reads := 0
	api := newStub(t, func(r stubRequest) (int, string) {
		switch {
		case r.Method == http.MethodPost && r.path == "/agent_gateways":
			_ = json.Unmarshal([]byte(r.body), &gw)
			gw["id"], gw["status"], gw["engine"], gw["gateway_url"] = "gw1", "pending", "litellm", nil
			return http.StatusCreated, jsonOf(map[string]any{"data": gw})
		case gw == nil:
			return http.StatusNotFound, notFound
		case r.Method == http.MethodGet && r.path == "/agent_gateways/gw1":
			if reads++; reads >= 2 {
				gw["status"], gw["gateway_url"] = "ready", "http://10.0.0.2:4000"
			}
			return http.StatusOK, jsonOf(map[string]any{"data": gw})
		case r.Method == http.MethodDelete && r.path == "/agent_gateways/gw1":
			gw = nil
			return http.StatusOK, `{"data":{}}`
		}
		return http.StatusNotFound, notFound
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if gw != nil {
				return fmt.Errorf("gateway left: %v", gw)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: api.config() + `
resource "telnyx_agent_gateway" "x" {
  name                   = "ci"
  wireguard_interface_id = "wg1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_agent_gateway.x", "status", "ready"),
					resource.TestCheckResourceAttr("telnyx_agent_gateway.x", "gateway_url", "http://10.0.0.2:4000"),
				),
			},
			{ResourceName: "telnyx_agent_gateway.x", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"timeouts"}},
		},
	})
}

// A gateway that fails to start fails the apply but stays in state, so the
// next apply replaces it.
func TestAgentGatewayFailed(t *testing.T) {
	defer func(d time.Duration) { agentGatewayPoll = d }(agentGatewayPoll)
	agentGatewayPoll = 10 * time.Millisecond

	var gw map[string]any
	api := newStub(t, func(r stubRequest) (int, string) {
		switch {
		case r.Method == http.MethodPost:
			_ = json.Unmarshal([]byte(r.body), &gw)
			gw["id"], gw["status"] = "gw1", "failed"
			return http.StatusCreated, jsonOf(map[string]any{"data": gw})
		case gw == nil:
			return http.StatusNotFound, notFound
		case r.Method == http.MethodGet:
			return http.StatusOK, jsonOf(map[string]any{"data": gw})
		case r.Method == http.MethodDelete:
			gw = nil
			return http.StatusOK, `{"data":{}}`
		}
		return http.StatusNotFound, notFound
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		Steps: []resource.TestStep{{
			Config: api.config() + `
resource "telnyx_agent_gateway" "x" {
  wireguard_interface_id = "wg1"
}`,
			ExpectError: regexp.MustCompile(`agent gateway gw1 is failed`),
		}},
	})
}

// Tool permissions are replaced by PUT, and destroy grants none.
func TestAgentGatewayKeyToolPermissions(t *testing.T) {
	var grant any
	api := newStub(t, func(r stubRequest) (int, string) {
		const p = "/agent_gateways/gw1/keys/k1/tool_permissions"
		switch {
		case r.path != p:
			return http.StatusNotFound, notFound
		case r.Method == http.MethodPut:
			var in map[string]any
			_ = json.Unmarshal([]byte(r.body), &in)
			grant = in["servers"]
			return http.StatusOK, jsonOf(map[string]any{"data": in})
		case r.Method == http.MethodGet:
			return http.StatusOK, jsonOf(map[string]any{"data": map[string]any{"servers": grant}})
		}
		return http.StatusNotFound, notFound
	})
	cfg := func(tools string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_agent_gateway_key_tool_permissions" "x" {
  gateway_id = "gw1"
  key_id     = "k1"
  servers = [
    { id = "s1", tools = %s },
    { id = "s2", tools = [] },
  ]
}`, tools)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			return api.sent(`PUT /agent_gateways/gw1/keys/k1/tool_permissions {"servers":[]}`)
		},
		Steps: []resource.TestStep{
			{Config: cfg(`["search"]`), Check: resource.TestCheckResourceAttr("telnyx_agent_gateway_key_tool_permissions.x", "servers.#", "2")},
			{Config: cfg(`["search", "fetch"]`)},
			{
				ResourceName: "telnyx_agent_gateway_key_tool_permissions.x", ImportState: true, ImportStateVerify: true,
				ImportStateId: "gw1/k1", ImportStateVerifyIdentifierAttribute: "key_id",
			},
		},
	})
}
