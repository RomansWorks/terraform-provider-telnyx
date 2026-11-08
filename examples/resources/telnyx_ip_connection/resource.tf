resource "telnyx_ip_connection" "example" {
  connection_name    = "Office PBX"
  transport_protocol = "UDP"
  tags               = ["pbx"]

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
  }
}
