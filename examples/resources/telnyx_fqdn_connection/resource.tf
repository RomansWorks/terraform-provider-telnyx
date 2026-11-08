resource "telnyx_fqdn_connection" "example" {
  connection_name    = "PBX"
  transport_protocol = "UDP"
  tags               = ["pbx"]

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
  }
}
