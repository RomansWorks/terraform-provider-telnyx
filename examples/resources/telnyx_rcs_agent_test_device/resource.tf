resource "telnyx_rcs_agent_test_device" "qa_phone" {
  agent_id     = telnyx_rcs_agent.alerts.id
  phone_number = "+13125550142"
}
