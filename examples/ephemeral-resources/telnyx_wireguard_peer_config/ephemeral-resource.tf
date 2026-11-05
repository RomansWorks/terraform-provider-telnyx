# The WireGuard configuration template of a peer, never stored in state.
ephemeral "telnyx_wireguard_peer_config" "laptop" {
  wireguard_peer_id = telnyx_wireguard_peer.laptop.id
}
