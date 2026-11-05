variable "tailscale_auth_key" {
  type      = string
  sensitive = true
}

# The auth key is write-only: it is sent on create and never stored. Bump
# auth_key_wo_version to send a replacement after a failed enrollment.
resource "telnyx_tailscale_subnet_router" "office" {
  name              = "office"
  network_id        = telnyx_network.office.id
  region_code       = "ashburn-va"
  advertised_routes = ["10.10.0.0/24"]
  auth_key_wo       = var.tailscale_auth_key
}
