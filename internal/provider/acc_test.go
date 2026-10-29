// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// Acceptance tests create real objects in the Telnyx account behind
// TELNYX_API_KEY. Every object they create is named with accPrefix so the
// sweepers (go test ./internal/provider -sweep=all) can remove leftovers.

const accPrefix = "tf-acc-"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func accName() string { return accPrefix + acctest.RandString(8) }

func accClient() (*client.Client, error) {
	base := os.Getenv("TELNYX_BASE_URL")
	return client.New(client.Config{APIKey: os.Getenv("TELNYX_API_KEY"), BaseURL: base, MaxRetries: client.DefaultMaxRetries})
}

// accList returns the objects of a list endpoint, following page numbers.
func accList(ctx context.Context, c *client.Client, path string) ([]map[string]any, error) {
	var out []map[string]any
	for page := 1; page <= 50; page++ {
		q := url.Values{"page[number]": {fmt.Sprint(page)}, "page[size]": {"100"}}
		data, err := c.Send(ctx, http.MethodGet, path, q, "", nil)
		if err != nil {
			return nil, err
		}
		var items []map[string]any
		var wrapped struct {
			Data []map[string]any `json:"data"`
		}
		if json.Unmarshal(data, &items) != nil {
			if err := json.Unmarshal(data, &wrapped); err != nil {
				return nil, err
			}
			items = wrapped.Data
		}
		out = append(out, items...)
		if len(items) < 100 {
			return out, nil
		}
	}
	return out, nil
}

// addSweeper deletes objects of a list endpoint whose name field starts
// with accPrefix.
func addSweeper(name, list, nameField, deletePath string) {
	resource.AddTestSweepers(name, &resource.Sweeper{
		Name: name,
		F: func(string) error {
			ctx := context.Background()
			c, err := accClient()
			if err != nil {
				return err
			}
			items, err := accList(ctx, c, list)
			if err != nil {
				return err
			}
			for _, it := range items {
				n, _ := it[nameField].(string)
				id, _ := it["id"].(string)
				if !strings.HasPrefix(n, accPrefix) || id == "" {
					continue
				}
				if _, err := c.Send(ctx, http.MethodDelete, strings.ReplaceAll(deletePath, "{id}", url.PathEscape(id)), nil, "", nil); err != nil && !client.IsNotFound(err) {
					return fmt.Errorf("deleting %s %s: %w", name, id, err)
				}
			}
			return nil
		},
	})
}

func init() {
	addSweeper("telnyx_ai_assistant", "/ai/assistants", "name", "/ai/assistants/{id}")
	addSweeper("telnyx_ai_tool", "/ai/tools", "display_name", "/ai/tools/{id}")
	addSweeper("telnyx_ai_mcp_server", "/ai/mcp_servers", "name", "/ai/mcp_servers/{id}")
	addSweeper("telnyx_outbound_voice_profile", "/outbound_voice_profiles", "name", "/outbound_voice_profiles/{id}")
	// Numbers bought by the number order test carry accPrefix as their
	// customer reference. Avatars cannot be deleted through the API.
	addSweeper("telnyx_number_order", "/phone_numbers", "customer_reference", "/phone_numbers/{id}")
}

// accDestroyed checks that every object of a type in state is gone.
func accDestroyed(typeName, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := accClient()
		if err != nil {
			return err
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != typeName {
				continue
			}
			_, err := c.Send(context.Background(), http.MethodGet, strings.ReplaceAll(path, "{id}", url.PathEscape(rs.Primary.ID)), nil, "", nil)
			if err == nil {
				return fmt.Errorf("%s %s still exists", typeName, rs.Primary.ID)
			}
			if !client.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func TestAccAIMCPServer(t *testing.T) {
	name := accName()
	cfg := func(tools string) string {
		return fmt.Sprintf(`
resource "telnyx_ai_mcp_server" "x" {
  name          = %q
  type          = "http"
  url           = "https://mcp.example.com/mcp"
  allowed_tools = [%s]
}
data "telnyx_ai_mcp_server" "x" {
  id = telnyx_ai_mcp_server.x.id
}`, name, tools)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             accDestroyed("telnyx_ai_mcp_server", "/ai/mcp_servers/{id}"),
		Steps: []resource.TestStep{
			{Config: cfg(`"search"`), Check: resource.TestCheckResourceAttrPair("data.telnyx_ai_mcp_server.x", "name", "telnyx_ai_mcp_server.x", "name")},
			{Config: cfg(`"search", "fetch"`), Check: resource.TestCheckResourceAttr("telnyx_ai_mcp_server.x", "allowed_tools.#", "2")},
			{ResourceName: "telnyx_ai_mcp_server.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccOutboundVoiceProfile(t *testing.T) {
	name := accName()
	cfg := func(limit string) string {
		return fmt.Sprintf(`
resource "telnyx_outbound_voice_profile" "x" {
  name                      = %q
  daily_spend_limit         = %q
  daily_spend_limit_enabled = true
  tags                      = ["terraform"]
}`, name, limit)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             accDestroyed("telnyx_outbound_voice_profile", "/outbound_voice_profiles/{id}"),
		Steps: []resource.TestStep{
			{Config: cfg("10.00")},
			{Config: cfg("20.00"), Check: resource.TestCheckResourceAttr("telnyx_outbound_voice_profile.x", "daily_spend_limit", "20.00")},
			{ResourceName: "telnyx_outbound_voice_profile.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccAITool(t *testing.T) {
	name := accName()
	cfg := func(url string) string {
		return fmt.Sprintf(`
resource "telnyx_ai_tool" "x" {
  type         = "webhook"
  display_name = %q
  webhook = jsonencode({
    name        = "order_status"
    description = "Look up an order."
    url         = %q
    method      = "POST"
  })
}`, name, url)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             accDestroyed("telnyx_ai_tool", "/ai/tools/{id}"),
		Steps: []resource.TestStep{
			{Config: cfg("https://example.com/orders")},
			{Config: cfg("https://example.com/orders/v2")},
			{ResourceName: "telnyx_ai_tool.x", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"webhook"}},
		},
	})
}

func TestAccAIAssistant(t *testing.T) {
	name := accName()
	cfg := func(greeting string) string {
		return fmt.Sprintf(`
resource "telnyx_ai_assistant" "x" {
  name         = %q
  instructions = "You answer questions about opening hours."
  greeting     = %q
}
data "telnyx_ai_assistant" "x" {
  id = telnyx_ai_assistant.x.id
}`, name, greeting)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             accDestroyed("telnyx_ai_assistant", "/ai/assistants/{id}"),
		Steps: []resource.TestStep{
			{Config: cfg("Hello"), Check: resource.TestCheckResourceAttrSet("data.telnyx_ai_assistant.x", "model")},
			{Config: cfg("Hi there")},
			{ResourceName: "telnyx_ai_assistant.x", ImportState: true, ImportStateVerify: true},
		},
	})
}

// Avatars cannot be deleted, so each run leaves one avatar in the account.
func TestAccAIAvatar(t *testing.T) {
	img, err := filepath.Abs("testdata/avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	name := accName()
	cfg := func(n string) string {
		return fmt.Sprintf(`
resource "telnyx_ai_avatar" "x" {
  name       = %q
  image_path = %q
}`, n, img)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg(name), Check: resource.TestCheckResourceAttr("telnyx_ai_avatar.x", "avatar_type", "custom")},
			{Config: cfg(name + "-renamed")},
			{ResourceName: "telnyx_ai_avatar.x", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"image_path", "image_sha256", "image_url", "image_url_expires_at"}},
		},
	})
}

// Needs a phone number in the test account: TELNYX_ACC_PHONE_NUMBER_ID.
// The test changes the number's inbound call screening and leaves it
// disabled.
func TestAccPhoneNumberVoiceSettings(t *testing.T) {
	id := os.Getenv("TELNYX_ACC_PHONE_NUMBER_ID")
	if id == "" {
		t.Skip("TELNYX_ACC_PHONE_NUMBER_ID is not set")
	}
	cfg := func(screening string) string {
		return fmt.Sprintf(`
resource "telnyx_phone_number_voice_settings" "x" {
  phone_number_id        = %q
  inbound_call_screening = %q
}`, id, screening)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg("flag_calls")},
			{
				ResourceName: "telnyx_phone_number_voice_settings.x", ImportState: true, ImportStateVerify: true,
				ImportStateId: id, ImportStateVerifyIdentifierAttribute: "phone_number_id",
			},
			{Config: cfg("disabled")},
		},
	})
}

// Buys a real phone number. Runs only with TELNYX_ACC_BILLABLE=1; the
// number is released when the test ends, and the sweeper releases any
// number whose customer reference starts with accPrefix.
func TestAccNumberOrder(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" || os.Getenv("TELNYX_ACC_BILLABLE") != "1" {
		t.Skip("buys a phone number; set TELNYX_ACC_BILLABLE=1 to run")
	}
	testAccPreCheck(t)
	ctx := context.Background()
	c, err := accClient()
	if err != nil {
		t.Fatal(err)
	}
	q := url.Values{"filter[country_code]": {"US"}, "filter[features]": {"voice"}, "filter[limit]": {"1"}}
	data, err := c.Send(ctx, http.MethodGet, "/available_phone_numbers", q, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var avail struct {
		Data []struct {
			PhoneNumber string `json:"phone_number"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &avail); err != nil || len(avail.Data) == 0 {
		t.Fatalf("no available number to order: %v", err)
	}
	number := avail.Data[0].PhoneNumber
	ref := accName()
	t.Cleanup(func() { releaseNumber(t, c, number) })
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "telnyx_number_order" "x" {
  customer_reference = %q
  phone_numbers      = [{ phone_number = %q }]
}`, ref, number),
				Check: resource.TestCheckResourceAttr("telnyx_number_order.x", "status", "success"),
			},
			{ResourceName: "telnyx_number_order.x", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"timeouts"}},
		},
	})
}

func releaseNumber(t *testing.T, c *client.Client, number string) {
	ctx := context.Background()
	q := url.Values{"filter[phone_number]": {number}}
	data, err := c.Send(ctx, http.MethodGet, "/phone_numbers", q, "", nil)
	if err != nil {
		t.Errorf("finding %s to release it: %v", number, err)
		return
	}
	var found struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &found); err != nil {
		t.Errorf("finding %s to release it: %v", number, err)
		return
	}
	for _, n := range found.Data {
		if _, err := c.Send(ctx, http.MethodDelete, "/phone_numbers/"+url.PathEscape(n.ID), nil, "", nil); err != nil && !client.IsNotFound(err) {
			t.Errorf("releasing %s: %v; release it in the Telnyx portal", number, err)
		}
	}
}
