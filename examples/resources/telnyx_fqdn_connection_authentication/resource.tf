# Destroy removes the settings from Terraform state only.
resource "telnyx_fqdn_connection_authentication" "example" {
  fqdn_connection_id           = telnyx_fqdn_connection.example.id
  fqdn_outbound_authentication = "credential-authentication"
  user_name                    = "acme-pbx"
  password                     = var.pbx_password
}

variable "pbx_password" {
  type      = string
  sensitive = true
}
