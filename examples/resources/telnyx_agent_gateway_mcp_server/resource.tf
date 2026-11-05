resource "telnyx_agent_gateway_mcp_server" "docs" {
  gateway_id = telnyx_agent_gateway.example.id
  name       = "docs"
  url        = "https://mcp.example.com/mcp"
  transport  = "http"

  auth_wo = jsonencode({
    type  = "bearer_token"
    value = var.docs_mcp_token
  })
  auth_wo_version = 1
}

variable "docs_mcp_token" {
  type      = string
  sensitive = true
}
