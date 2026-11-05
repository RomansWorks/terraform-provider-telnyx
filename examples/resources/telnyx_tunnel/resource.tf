resource "telnyx_tunnel" "lab" {
  name = "lab"
}

# Pass the token to the host that runs `telnyx-edge tunnel run --token`.
output "lab_tunnel_token" {
  value     = telnyx_tunnel.lab.tunnel_token
  sensitive = true
}
