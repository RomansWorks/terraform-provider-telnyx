// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// The phone number actions send their operation once; a bundle status
// change without bundle_id removes the number from its bundle with null.
func TestPhoneNumberActions(t *testing.T) {
	api := newStub(t, func(r stubRequest) (int, string) {
		switch r.path {
		case "/phone_numbers/pn1/actions/bundle_status_change", "/phone_numbers/pn1/actions/enable_emergency":
			return http.StatusOK, `{"data":{"id":"pn1"}}`
		case "/phone_numbers/actions/verify_ownership":
			return http.StatusOK, `{"data":{"found":[],"not_found":["+15551234568"]}}`
		case "/phone_number_blocks/jobs/delete_phone_number_block":
			return http.StatusAccepted, `{"data":{"id":"j1","status":"pending"}}`
		}
		return http.StatusNotFound, notFound
	})
	config := api.config() + `
action "telnyx_phone_number_bundle_status_change" "remove" {
  config {
    id = "pn1"
  }
}
action "telnyx_phone_number_bundle_status_change" "assign" {
  config {
    id        = "pn1"
    bundle_id = "b1"
  }
}
action "telnyx_phone_number_enable_emergency" "e" {
  config {
    id                   = "pn1"
    emergency_enabled    = true
    emergency_address_id = "addr1"
  }
}
action "telnyx_phone_number_verify_ownership" "v" {
  config {
    phone_numbers = ["+15551234568"]
  }
}
action "telnyx_phone_number_block_delete" "d" {
  config {
    phone_number_block_id = "blk1"
  }
}
resource "terraform_data" "t" {
  input = "x"
  lifecycle {
    action_trigger {
      events = [after_create]
      actions = [
        action.telnyx_phone_number_bundle_status_change.remove,
        action.telnyx_phone_number_bundle_status_change.assign,
        action.telnyx_phone_number_enable_emergency.e,
        action.telnyx_phone_number_verify_ownership.v,
        action.telnyx_phone_number_block_delete.d,
      ]
    }
  }
}`
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		ProtoV6ProviderFactories: stubFactories(),
		Steps: []resource.TestStep{{
			Config: config,
			Check: func(*terraform.State) error {
				for _, want := range []string{
					`PATCH /phone_numbers/pn1/actions/bundle_status_change {"bundle_id":null}`,
					`PATCH /phone_numbers/pn1/actions/bundle_status_change {"bundle_id":"b1"}`,
					`POST /phone_numbers/pn1/actions/enable_emergency {"emergency_address_id":"addr1","emergency_enabled":true}`,
					`POST /phone_numbers/actions/verify_ownership {"phone_numbers":["+15551234568"]}`,
					`POST /phone_number_blocks/jobs/delete_phone_number_block {"phone_number_block_id":"blk1"}`,
				} {
					if err := api.sent(want); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// Job lookups read a job by id, keep the saved external PIN sensitive and
// take numeric identifiers among the job's target numbers.
func TestPhoneNumberJobLookups(t *testing.T) {
	api := newStub(t, func(r stubRequest) (int, string) {
		switch r.path {
		case "/phone_numbers/jobs/j1":
			return http.StatusOK, `{"data":{"id":"j1","type":"update_phone_numbers","status":"completed",
				"parameters":{"phone_numbers":["+15551234567",1293384261075731499],"external_pin":"1234"}}}`
		case "/phone_number_blocks/jobs/j2":
			return http.StatusOK, `{"data":{"id":"j2","type":"delete_phone_number_block","status":"in_progress"}}`
		case "/phone_numbers/slim":
			return http.StatusOK, `{"data":[{"id":"pn1","phone_number":"+15551234567"}],"meta":{"total_pages":1}}`
		case "/phone_numbers/voice":
			return http.StatusOK, `{"data":[{"id":"pn1","phone_number":"+15551234567"}],"meta":{"total_pages":1}}`
		case "/phone_numbers/regulatory_requirements":
			return http.StatusOK, `{"data":[{"phone_number":"+41215470622","phone_number_type":"local"}],"meta":{"total_pages":1}}`
		}
		return http.StatusNotFound, notFound
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		Steps: []resource.TestStep{{
			Config: api.config() + `
data "telnyx_phone_number_job" "j" {
  id = "j1"
}
data "telnyx_phone_number_block_job" "b" {
  id = "j2"
}
data "telnyx_phone_numbers_slim" "s" {
  include_tags = true
}
data "telnyx_phone_numbers_voice_settings" "v" {}
data "telnyx_phone_number_regulatory_requirements" "r" {
  filter_phone_number = "+41215470622"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.telnyx_phone_number_job.j", "status", "completed"),
				resource.TestCheckResourceAttr("data.telnyx_phone_number_job.j", "parameters.phone_numbers.1", "1293384261075731499"),
				resource.TestCheckResourceAttr("data.telnyx_phone_number_job.j", "parameters.external_pin", "1234"),
				resource.TestCheckResourceAttr("data.telnyx_phone_number_block_job.b", "status", "in_progress"),
				resource.TestCheckResourceAttr("data.telnyx_phone_numbers_slim.s", "phone_numbers_slim.0.id", "pn1"),
				resource.TestCheckResourceAttr("data.telnyx_phone_numbers_voice_settings.v", "phone_numbers_voice_settings.0.id", "pn1"),
				resource.TestCheckResourceAttr("data.telnyx_phone_number_regulatory_requirements.r", "phone_number_regulatory_requirements.0.phone_number_type", "local"),
				func(*terraform.State) error {
					return api.sent(`GET /phone_numbers/slim?include_tags=true&page%5Bnumber%5D=1&page%5Bsize%5D=250`)
				},
			),
		}},
	})
}
