resource "telnyx_fax_application" "example" {
  application_name    = "Office fax"
  webhook_event_url   = "https://example.com/webhooks/fax"
  fax_email_recipient = "fax@example.com"
  tags                = ["office"]

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
  }
}
