resource "telnyx_agent_gateway_team" "support" {
  gateway_id          = telnyx_agent_gateway.example.id
  name                = "support"
  max_budget          = 100
  budget_resets_every = "30d"
}

resource "telnyx_agent_gateway_key" "support_bot" {
  gateway_id          = telnyx_agent_gateway.example.id
  name                = "support-bot"
  team_id             = telnyx_agent_gateway_team.support.id
  models              = ["gpt-4o-mini"]
  requests_per_minute = 60
}

# The key's value is only available after create.
output "support_bot_key" {
  value     = telnyx_agent_gateway_key.support_bot.key
  sensitive = true
}
