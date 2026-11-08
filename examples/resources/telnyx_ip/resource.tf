resource "telnyx_ip" "example" {
  connection_id = telnyx_ip_connection.example.id
  ip_address    = "203.0.113.20"
  port          = 5060
}
