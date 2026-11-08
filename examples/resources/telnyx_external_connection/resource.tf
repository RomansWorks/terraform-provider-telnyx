resource "telnyx_external_connection" "example" {
  external_sip_connection = "zoom"
  active                  = true
  tags                    = ["zoom"]

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
    channel_limit             = 10
  }
}
