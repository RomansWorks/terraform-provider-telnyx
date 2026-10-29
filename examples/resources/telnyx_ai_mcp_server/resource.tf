resource "telnyx_ai_mcp_server" "docs" {
  name          = "docs"
  type          = "http"
  url           = "https://mcp.example.com/mcp"
  allowed_tools = ["search", "fetch"]
}
