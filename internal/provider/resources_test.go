// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// These tests run the pilot resources through Terraform against a fake API
// that follows the pinned spec's shapes. They need no credentials. Every
// step also checks that the plan after apply is empty.

func TestAIMCPServer(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/mcp_servers": {}})
	cfg := func(url string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_mcp_server" "x" {
  name          = "docs"
  type          = "http"
  url           = %q
  allowed_tools = ["search"]
}
data "telnyx_ai_mcp_server" "x" {
  id = telnyx_ai_mcp_server.x.id
}`, url)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy:             func(*terraform.State) error { return emptyAPI(api) },
		Steps: []resource.TestStep{
			{
				Config: cfg("https://mcp.example.com/a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_mcp_server.x", "id", "1"),
					resource.TestCheckResourceAttr("telnyx_ai_mcp_server.x", "created_at", "2026-10-29T00:00:00Z"),
					resource.TestCheckResourceAttr("data.telnyx_ai_mcp_server.x", "url", "https://mcp.example.com/a"),
				),
			},
			{
				// The update is a PUT, so the whole object is sent.
				Config: cfg("https://mcp.example.com/b"),
				Check: func(*terraform.State) error {
					return api.sent(`PUT /ai/mcp_servers/1 {"allowed_tools":["search"],"created_at":"2026-10-29T00:00:00Z","name":"docs","type":"http","url":"https://mcp.example.com/b"}`)
				},
			},
			{ResourceName: "telnyx_ai_mcp_server.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestOutboundVoiceProfile(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/outbound_voice_profiles": {
		envelope: true,
		defaults: map[string]any{
			"record_type": "outbound_voice_profile", "connections_count": 0, "enabled": true,
			"traffic_type": "conversational", "service_plan": "global", "usage_payment_method": "rate-deck",
			"concurrent_call_limit": nil, "tags": []any{}, "whitelisted_destinations": []any{"US", "CA"},
			"daily_spend_limit_enabled": false, "max_destination_rate": 0.5,
			"call_recording": map[string]any{
				"call_recording_type": "none", "call_recording_caller_phone_numbers": []any{},
				"call_recording_channels": "single", "call_recording_format": "wav",
			},
		},
	}})
	cfg := func(limit string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_outbound_voice_profile" "x" {
  name              = "office"
  daily_spend_limit = %q
  call_recording = {
    call_recording_type = "all"
  }
}`, limit)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy:             func(*terraform.State) error { return emptyAPI(api) },
		Steps: []resource.TestStep{
			{
				Config: cfg("100.00"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_outbound_voice_profile.x", "enabled", "true"),
					resource.TestCheckResourceAttr("telnyx_outbound_voice_profile.x", "whitelisted_destinations.#", "2"),
					resource.TestCheckResourceAttr("telnyx_outbound_voice_profile.x", "call_recording.call_recording_type", "all"),
				),
			},
			{
				// PATCH sends only what changed.
				Config: cfg("150.00"),
				Check: func(*terraform.State) error {
					return api.sent(`PATCH /outbound_voice_profiles/1 {"daily_spend_limit":"150.00"}`)
				},
			},
			{
				// Drift made in the portal is planned back.
				PreConfig: func() { api.get("/outbound_voice_profiles/1")["name"] = "renamed" },
				Config:    cfg("150.00"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_outbound_voice_profile.x", plancheck.ResourceActionUpdate)},
				},
			},
			{ResourceName: "telnyx_outbound_voice_profile.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAITool(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/tools": {
		defaults: map[string]any{"timeout_ms": 5000},
		created: func(obj map[string]any) {
			// The API returns the variant's settings, with its defaults, as
			// tool_definition, and not under the variant's name.
			def := map[string]any{"method": "POST"}
			for k, v := range obj["webhook"].(map[string]any) { //nolint:forcetypeassert
				def[k] = v
			}
			obj["tool_definition"] = def
			delete(obj, "webhook")
		},
	}})
	cfg := func(url string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_tool" "x" {
  type         = "webhook"
  display_name = "Order status"
  webhook      = jsonencode({ name = "order_status", url = %q })
}`, url)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy:             func(*terraform.State) error { return emptyAPI(api) },
		Steps: []resource.TestStep{
			{
				Config: cfg("https://example.com/a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_tool.x", "timeout_ms", "5000"),
					resource.TestCheckResourceAttr("telnyx_ai_tool.x", "tool_definition", `{"method":"POST","name":"order_status","url":"https://example.com/a"}`),
					resource.TestCheckNoResourceAttr("telnyx_ai_tool.x", "function"),
				),
			},
			{
				// A webhook URL changed in the portal is drift.
				PreConfig: func() {
					api.get("/ai/tools/1")["tool_definition"].(map[string]any)["url"] = "https://example.com/changed" //nolint:forcetypeassert
				},
				Config: cfg("https://example.com/a"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_ai_tool.x", plancheck.ResourceActionUpdate)},
				},
			},
			{
				// Import reads the settings back from tool_definition, with the
				// server's defaults included.
				ResourceName: "telnyx_ai_tool.x", ImportState: true, ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

func TestAIAssistant(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/assistants": {
		defaults: map[string]any{
			"model":            "moonshotai/Kimi-K2.5",
			"enabled_features": []any{"telephony"},
			"version_id":       "main",
			"voice_settings":   map[string]any{"voice": "Telnyx.NaturalHD.astra", "voice_speed": 1},
			"tools":            []any{},
		},
	}})
	cfg := func(greeting string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_assistant" "x" {
  name         = "Front desk"
  instructions = "Answer questions about opening hours."
  greeting     = %q
  tools = [{
    type    = "hangup"
    hangup  = {}
  }]
}`, greeting)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy:             func(*terraform.State) error { return emptyAPI(api) },
		Steps: []resource.TestStep{
			{
				Config: cfg("Hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_assistant.x", "model", "moonshotai/Kimi-K2.5"),
					resource.TestCheckResourceAttr("telnyx_ai_assistant.x", "voice_settings.voice", "Telnyx.NaturalHD.astra"),
					resource.TestCheckResourceAttr("telnyx_ai_assistant.x", "tools.0.type", "hangup"),
				),
			},
			{
				// Assistants are updated with POST, sending only changes.
				Config: cfg("Hi there"),
				Check: func(*terraform.State) error {
					return api.sent(`POST /ai/assistants/1 {"greeting":"Hi there"}`)
				},
			},
			{ResourceName: "telnyx_ai_assistant.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAIAvatar(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/avatars": {
		envelope: true,
		created: func(obj map[string]any) {
			obj["provider"], obj["avatar_type"], obj["hosted"] = "telnyx", "custom", true
			obj["image_url"] = "https://cdn.example.com/" + obj["id"].(string) + ".png" //nolint:forcetypeassert
		},
	}})
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	write := func(data string) {
		if err := os.WriteFile(img, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("first image")
	cfgAt := func(name, path string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_avatar" "x" {
  name       = %q
  image_path = %q
}`, name, path)
	}
	cfg := func(name string) string { return cfgAt(name, img) }
	moved := filepath.Join(dir, "moved.png")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy: func(*terraform.State) error {
			if api.sentPrefix("DELETE") {
				return fmt.Errorf("avatars cannot be deleted; no DELETE may be sent")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("Portrait"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_avatar.x", "avatar_type", "custom"),
					resource.TestCheckResourceAttr("telnyx_ai_avatar.x", "image_sha256", "fb4516b50946085bd15d6807d1526c967afdfae509a6088b7817853bac9bed0c"),
					func(*terraform.State) error {
						return api.sent(`POST /ai/avatars multipart name=Portrait image_file=file(a.png,image/png,11 bytes)`)
					},
				),
			},
			{
				// A new name is a PATCH.
				Config: cfg("Portrait 2"),
				Check: func(*terraform.State) error {
					return api.sent(`PATCH /ai/avatars/1 {"name":"Portrait 2"}`)
				},
			},
			{
				// The same image at another path changes nothing in Telnyx.
				PreConfig: func() {
					data, _ := os.ReadFile(img)
					_ = os.WriteFile(moved, data, 0o600)
				},
				Config: cfgAt("Portrait 2", moved),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_ai_avatar.x", plancheck.ResourceActionUpdate)},
				},
				Check: func(*terraform.State) error {
					if api.sentPrefix("PATCH /ai/avatars/1 {}") {
						return fmt.Errorf("a path-only change must not call the API")
					}
					return nil
				},
			},
			{
				// A new image is a new avatar.
				PreConfig: func() { write("second image") },
				Config:    cfg("Portrait 2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_ai_avatar.x", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.TestCheckResourceAttr("telnyx_ai_avatar.x", "id", "2"),
			},
			{
				ResourceName: "telnyx_ai_avatar.x", ImportState: true, ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"image_path", "image_sha256"},
			},
		},
	})
}

func TestAIAvatarRejectsUnknownImageType(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/avatars": {envelope: true}})
	img := filepath.Join(t.TempDir(), "a.gif")
	if err := os.WriteFile(img, []byte("gif"), 0o600); err != nil {
		t.Fatal(err)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		Steps: []resource.TestStep{{
			Config: api.config() + fmt.Sprintf(`
resource "telnyx_ai_avatar" "x" {
  name       = "x"
  image_path = %q
}`, img),
			ExpectError: regexp.MustCompile(`(?s)must\s+be\s+a\s+\.jpg`),
		}},
	})
}

func TestPhoneNumberVoiceSettings(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/phone_numbers": {envelope: true}})
	api.seed("/phone_numbers/123/voice", map[string]any{
		"id": "123", "record_type": "voice_settings", "phone_number": "+13125550100",
		"tech_prefix_enabled": false, "translated_number": "", "usage_payment_method": "pay-per-minute",
		"inbound_call_screening": "disabled",
		"call_forwarding":        map[string]any{"call_forwarding_enabled": false, "forwards_to": "", "forwarding_type": "always"},
	})
	cfg := func(screening string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_phone_number_voice_settings" "x" {
  phone_number_id        = "123"
  inbound_call_screening = %q
}
data "telnyx_phone_number_voice_settings" "x" {
  phone_number_id = telnyx_phone_number_voice_settings.x.phone_number_id
}`, screening)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy: func(*terraform.State) error {
			if api.get("/phone_numbers/123/voice") == nil {
				return fmt.Errorf("destroy must leave the phone number's settings")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				// Create adopts the existing settings with a PATCH.
				Config: cfg("flag_calls"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_phone_number_voice_settings.x", "phone_number", "+13125550100"),
					resource.TestCheckResourceAttr("telnyx_phone_number_voice_settings.x", "call_forwarding.forwarding_type", "always"),
					resource.TestCheckResourceAttr("data.telnyx_phone_number_voice_settings.x", "inbound_call_screening", "flag_calls"),
					func(*terraform.State) error {
						return api.sent(`PATCH /phone_numbers/123/voice {"inbound_call_screening":"flag_calls"}`)
					},
				),
			},
			{
				ResourceName: "telnyx_phone_number_voice_settings.x", ImportState: true, ImportStateVerify: true,
				ImportStateId: "123", ImportStateVerifyIdentifierAttribute: "phone_number_id",
			},
		},
	})
}

func TestNumberOrder(t *testing.T) {
	defer func(d time.Duration) { numberOrderPoll = d }(numberOrderPoll)
	numberOrderPoll = 10 * time.Millisecond
	api := newFakeTelnyx(t, map[string]*collection{"/number_orders": {
		envelope: true,
		created: func(obj map[string]any) {
			obj["status"], obj["phone_numbers_count"], obj["requirements_met"] = "pending", 1, true
			for _, n := range obj["phone_numbers"].([]any) { //nolint:forcetypeassert
				n := n.(map[string]any) //nolint:forcetypeassert
				n["id"], n["status"], n["record_type"] = "pn-1", "pending", "number_order_phone_number"
			}
		},
		read: func(obj map[string]any) {
			obj["status"] = "success" // done by the first poll
		},
	}})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		Steps: []resource.TestStep{
			{
				Config: api.config() + `
resource "telnyx_number_order" "x" {
  customer_reference = "terraform"
  phone_numbers      = [{ phone_number = "+13125550100" }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_number_order.x", "status", "success"),
					resource.TestCheckResourceAttr("telnyx_number_order.x", "phone_numbers.0.id", "pn-1"),
				),
			},
			{ResourceName: "telnyx_number_order.x", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"timeouts"}},
		},
	})
}

func TestNumberOrderFailure(t *testing.T) {
	defer func(d time.Duration) { numberOrderPoll = d }(numberOrderPoll)
	numberOrderPoll = 10 * time.Millisecond
	api := newFakeTelnyx(t, map[string]*collection{"/number_orders": {
		envelope: true,
		created:  func(obj map[string]any) { obj["status"] = "pending" },
		read:     func(obj map[string]any) { obj["status"] = "failure" },
	}})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		Steps: []resource.TestStep{{
			Config:      api.config() + `resource "telnyx_number_order" "x" { phone_numbers = [{ phone_number = "+13125550100" }] }`,
			ExpectError: regexp.MustCompile(`number\s+order\s+1\s+failed`),
		}},
	})
}

func emptyAPI(api *fakeTelnyx) error {
	if n := api.count(); n != 0 {
		return fmt.Errorf("%d objects left in the API after destroy", n)
	}
	return nil
}

func TestAIToolSettingsMustMatchType(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/ai/tools": {}})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		Steps: []resource.TestStep{
			{
				Config: api.config() + `
resource "telnyx_ai_tool" "x" {
  type         = "webhook"
  display_name = "x"
  handoff      = jsonencode({ ai_assistants = [] })
}`,
				ExpectError: regexp.MustCompile(`Tool settings do not match its type`),
			},
			{
				Config: api.config() + `
resource "telnyx_ai_tool" "x" {
  type         = "webhook"
  display_name = "x"
}`,
				ExpectError: regexp.MustCompile(`Exactly one of these attributes must be configured`),
			},
		},
	})
}

func TestNumberOrderDeleteWarningNamesNumbers(t *testing.T) {
	item := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"phone_number": tftypes.String}}
	state := engine.Values{"phone_numbers": tftypes.NewValue(tftypes.List{ElementType: item}, []tftypes.Value{
		tftypes.NewValue(item, map[string]tftypes.Value{"phone_number": tftypes.NewValue(tftypes.String, "+13125550100")}),
	})}
	if got := numberOrderHooks.DeleteWarning(state); !strings.Contains(got, "+13125550100") {
		t.Errorf("warning does not name the number: %s", got)
	}
}
