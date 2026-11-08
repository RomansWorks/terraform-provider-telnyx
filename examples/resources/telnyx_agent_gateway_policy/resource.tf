resource "telnyx_agent_gateway_policy" "default" {
  gateway_id  = telnyx_agent_gateway.example.id
  name        = "default"
  description = "Masks personal data in every request."
  guardrails = {
    add = [telnyx_agent_gateway_guardrail.pii.name]
  }
}

resource "telnyx_agent_gateway_policy_attachment" "everything" {
  gateway_id = telnyx_agent_gateway.example.id
  policy     = telnyx_agent_gateway_policy.default.name
}
