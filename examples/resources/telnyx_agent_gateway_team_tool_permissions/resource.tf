# Each apply replaces the team's whole grant. An empty tools list covers every
# tool of that server.
resource "telnyx_agent_gateway_team_tool_permissions" "support" {
  gateway_id = telnyx_agent_gateway.example.id
  team_id    = telnyx_agent_gateway_team.support.id
  servers = [
    { id = telnyx_agent_gateway_mcp_server.docs.id, tools = ["search"] },
    { id = telnyx_agent_gateway_api.billing.id, tools = [] },
  ]
}
