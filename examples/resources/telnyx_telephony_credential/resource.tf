resource "telnyx_telephony_credential" "web" {
  connection_id = telnyx_credential_connection.office.id
  name          = "web-client"
}
