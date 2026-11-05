# Revoke a leaked tunnel token: run with
#   terraform apply -invoke=action.telnyx_tunnel_rotate_token.lab
# The new token is not returned. Replace the tunnel to get a token in state.
action "telnyx_tunnel_rotate_token" "lab" {
  config {
    tunnel_id = telnyx_tunnel.lab.id
  }
}
