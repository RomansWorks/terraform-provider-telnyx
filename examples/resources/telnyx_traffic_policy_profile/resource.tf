# Let the SIM cards reach only the company's own domains and network.
resource "telnyx_traffic_policy_profile" "example" {
  type      = "whitelist"
  domains   = ["acme.example.com"]
  ip_ranges = ["203.0.113.0/24"]
}
