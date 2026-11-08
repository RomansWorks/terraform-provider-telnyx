resource "telnyx_network_route" "example" {
  network_id = telnyx_network.example.id
  cidr       = "10.20.0.0/16"
  next_hop = {
    type = "tunnel"
    id   = var.tunnel_id
  }
}

variable "tunnel_id" {
  type = string
}
