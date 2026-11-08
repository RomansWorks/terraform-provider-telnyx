# With a port the record type is "a"; without one it must be "srv".
resource "telnyx_fqdn" "example" {
  connection_id   = telnyx_fqdn_connection.example.id
  fqdn            = "sip.example.com"
  dns_record_type = "a"
  port            = 5060
}
