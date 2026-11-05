resource "telnyx_ai_gateway_token_group" "example" {
  name            = "internal-tools"
  allowed_models  = ["openai/gpt-4o-mini"]
  max_budget      = 100
  budget_duration = "30d"
  rpm_limit       = 60
}

resource "telnyx_ai_gateway_token_key" "example" {
  name           = "ci"
  token_group_id = telnyx_ai_gateway_token_group.example.id
}

# The key's token is only returned when it is created.
output "gateway_token" {
  value     = telnyx_ai_gateway_token_key.example.token
  sensitive = true
}
