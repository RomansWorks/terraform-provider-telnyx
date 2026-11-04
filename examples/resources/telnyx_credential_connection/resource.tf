variable "sip_password" {
  type      = string
  sensitive = true
}

resource "telnyx_credential_connection" "office" {
  connection_name = "office-pbx"
  user_name       = "officepbx"
  password        = var.sip_password
  tags            = ["office"]
}
