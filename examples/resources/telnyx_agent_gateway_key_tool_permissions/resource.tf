resource "telnyx_agent_gateway_key_tool_permissions" "support_bot" {
  gateway_id = telnyx_agent_gateway.example.id
  key_id     = telnyx_agent_gateway_key.support_bot.id
  servers = [
    { id = telnyx_agent_gateway_mcp_server.docs.id, tools = ["search"] },
  ]
}
