resource "telnyx_ai_test_agent" "front_desk" {
  agent_name   = "Front desk"
  assistant_id = telnyx_ai_assistant.front_desk.id
  description  = "Simulated callers for the front desk assistant."
  language     = "en"
  llm_model    = "openai/gpt-4o-mini"
}
