resource "telnyx_integration_secret" "example" {
  identifier = "crm_api"
  type       = "bearer"
  token      = var.crm_token
}

variable "crm_token" {
  type      = string
  sensitive = true
}
