# A WireGuard interface on the office network.
resource "telnyx_wireguard_interface" "office" {
  network_id  = telnyx_network.office.id
  region_code = "ashburn-va"
  name        = "office-wg"
}
