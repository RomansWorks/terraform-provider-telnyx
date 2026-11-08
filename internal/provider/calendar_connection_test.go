// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func seedCalendarConnection(api *fakeTelnyx) {
	api.seed("/calendar_connections/cc1", map[string]any{
		"id": "cc1", "platform": "google_calendar", "platform_email": "team@example.com",
		"status": "connected", "auto_join_policy": "per_meeting", "bot_name": "Telnyx",
		"created_at": "2026-11-08T00:00:00Z", "updated_at": "2026-11-08T00:00:00Z",
	})
}

// Create adopts a connection made by OAuth, and destroy leaves it connected.
func TestCalendarConnectionAdopt(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/calendar_connections": {envelope: true}})
	seedCalendarConnection(api)
	cfg := func(policy string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_calendar_connection" "x" {
  calendar_connection_id = "cc1"
  auto_join_policy       = %q
}`, policy)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy: func(*terraform.State) error {
			if api.get("/calendar_connections/cc1") == nil {
				return fmt.Errorf("destroy without release_on_destroy must leave the connection")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("all_meetings"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_calendar_connection.x", "platform_email", "team@example.com"),
					resource.TestCheckResourceAttr("telnyx_calendar_connection.x", "bot_name", "Telnyx"),
					func(*terraform.State) error {
						return api.sent(`PATCH /calendar_connections/cc1 {"auto_join_policy":"all_meetings"}`)
					},
				),
			},
			{
				ResourceName: "telnyx_calendar_connection.x", ImportState: true, ImportStateVerify: true,
				ImportStateId: "cc1", ImportStateVerifyIdentifierAttribute: "calendar_connection_id",
				ImportStateVerifyIgnore: []string{"release_on_destroy"},
			},
		},
	})
}

// With release_on_destroy, destroy disconnects the calendar.
func TestCalendarConnectionReleaseOnDestroy(t *testing.T) {
	api := newFakeTelnyx(t, map[string]*collection{"/calendar_connections": {envelope: true}})
	seedCalendarConnection(api)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: api.factories(),
		CheckDestroy: func(*terraform.State) error {
			if !api.sentPrefix("DELETE /calendar_connections/cc1") {
				return fmt.Errorf("destroy with release_on_destroy must send DELETE")
			}
			return nil
		},
		Steps: []resource.TestStep{{
			Config: api.config() + `
resource "telnyx_calendar_connection" "x" {
  calendar_connection_id = "cc1"
  bot_name               = "Notetaker"
  release_on_destroy     = true
}`,
			Check: resource.TestCheckResourceAttr("telnyx_calendar_connection.x", "bot_name", "Notetaker"),
		}},
	})
}
