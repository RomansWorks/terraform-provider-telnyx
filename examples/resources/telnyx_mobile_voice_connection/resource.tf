resource "telnyx_mobile_voice_connection" "example" {
  connection_name   = "Field team SIMs"
  active            = true
  webhook_event_url = "https://example.com/webhooks/mobile-voice"

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
  }
}
