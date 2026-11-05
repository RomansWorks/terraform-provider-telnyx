# Connects the SIM cards of a group to the office network.
resource "telnyx_private_wireless_gateway" "office" {
  name       = "office-pgw"
  network_id = telnyx_network.office.id
}
