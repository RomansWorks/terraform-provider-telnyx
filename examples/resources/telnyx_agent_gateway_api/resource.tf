resource "telnyx_agent_gateway_api" "billing" {
  gateway_id = telnyx_agent_gateway.example.id
  name       = "billing"
  spec_url   = "https://billing.example.com/openapi.json"
  base_url   = "https://billing.example.com/v1"
  tools      = ["getInvoice", "listInvoices"]

  auth_wo = jsonencode({
    type  = "bearer_token"
    value = var.billing_api_token
  })
  auth_wo_version = 1
}

variable "billing_api_token" {
  type      = string
  sensitive = true
}
