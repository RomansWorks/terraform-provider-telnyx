# Telnyx registers with your PBX as a SIP client.
resource "telnyx_uac_connection" "example" {
  connection_name = "Office PBX registration"
  tags            = ["pbx"]

  external_uac_settings = {
    proxy          = "pbx.example.com"
    username       = "telnyx"
    password       = var.pbx_password
    transport      = "UDP"
    expiration_sec = 3600
  }

  outbound = {
    outbound_voice_profile_id = telnyx_outbound_voice_profile.example.id
  }
}

variable "pbx_password" {
  type      = string
  sensitive = true
}
