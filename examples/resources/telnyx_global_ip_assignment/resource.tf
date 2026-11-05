# Deliver traffic for the Global IP over the laptop peer's tunnel.
resource "telnyx_global_ip_assignment" "edge" {
  global_ip_id      = telnyx_global_ip.edge.id
  wireguard_peer_id = telnyx_wireguard_peer.laptop.id
}
