# A peer with your own key pair. Without public_key, Telnyx generates a key
# pair and private_key holds the private key (sensitive, set on create only).
resource "telnyx_wireguard_peer" "laptop" {
  wireguard_interface_id = telnyx_wireguard_interface.office.id
  public_key             = "qF4EqlZq+5JL2IKYY8ij49daYyfKVhevJrcDxdqC8GU="
}

