resource "telnyx_oauth_client" "backend" {
  name                = "Backend"
  client_type         = "confidential"
  allowed_grant_types = ["client_credentials"]
  allowed_scopes      = ["admin"]
}

# The secret is returned only when the client is created.
output "backend_client_secret" {
  value     = telnyx_oauth_client.backend.client_secret
  sensitive = true
}
