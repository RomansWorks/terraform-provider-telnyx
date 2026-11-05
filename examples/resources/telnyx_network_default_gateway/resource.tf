# Send traffic without a more specific route through the laptop peer.
resource "telnyx_network_default_gateway" "office" {
  network_id        = telnyx_network.office.id
  wireguard_peer_id = telnyx_wireguard_peer.laptop.id
}
