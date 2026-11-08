resource "telnyx_ai_test_metric" "resolved" {
  name        = "Resolved"
  description = "Whether the assistant resolved the caller's request."
  type        = "llm_judge"
  eval_type   = "numeric"
  prompt      = "Score 1 if the assistant fully resolved the caller's request, otherwise 0."

  # Test agent ids are numbers.
  agents = [tonumber(telnyx_ai_test_agent.front_desk.id)]
}
