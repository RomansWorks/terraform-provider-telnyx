resource "telnyx_access_ip_range" "example" {
  cidr_block  = "203.0.113.0/24"
  description = "Office network"
}
