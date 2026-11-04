data "telnyx_sip_registration_status" "web" {
  credential_type = "telephony_credential"
  connection_id   = telnyx_credential_connection.office.id
  username        = telnyx_telephony_credential.web.sip_username
}
