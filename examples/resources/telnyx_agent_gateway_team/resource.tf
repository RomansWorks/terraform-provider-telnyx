resource "telnyx_agent_gateway_team" "support" {
  gateway_id          = telnyx_agent_gateway.example.id
  name                = "support"
  models              = ["gpt-4o-mini"]
  max_budget          = 100
  budget_resets_every = "30d"
  requests_per_minute = 120
}
