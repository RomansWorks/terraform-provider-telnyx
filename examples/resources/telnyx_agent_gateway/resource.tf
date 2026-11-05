resource "telnyx_agent_gateway" "example" {
  name                   = "agents"
  wireguard_interface_id = telnyx_wireguard_interface.example.id

  timeouts = {
    create = "20m"
  }
}
