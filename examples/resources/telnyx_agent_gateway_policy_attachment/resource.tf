# Apply the policy to the requests for the listed models only. Every change
# replaces the attachment.
resource "telnyx_agent_gateway_policy_attachment" "cheap_models" {
  gateway_id = telnyx_agent_gateway.example.id
  policy     = telnyx_agent_gateway_policy.default.name
  models     = ["gpt-4o-mini", "claude-*"]
  priority   = 10
}
