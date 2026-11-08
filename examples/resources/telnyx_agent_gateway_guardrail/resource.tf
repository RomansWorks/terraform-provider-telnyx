resource "telnyx_agent_gateway_guardrail" "pii" {
  gateway_id = telnyx_agent_gateway.example.id
  name       = "pii"
  type       = "content_filter"
  modes      = ["before_request"]

  settings_wo = jsonencode({
    patterns = [{ name = "email", action = "mask" }]
  })
  settings_wo_version = 1
}
