resource "telnyx_ai_test_scenario" "book_appointment" {
  name                    = "Book an appointment"
  personality             = 3
  agent                   = tonumber(telnyx_ai_test_agent.front_desk.id)
  first_message           = "Hi, I'd like to book a haircut for Friday."
  instructions            = "You want a haircut on Friday afternoon and accept any time after 2 pm."
  expected_outcome_prompt = "The assistant books an appointment on Friday after 2 pm."
  max_duration            = 300
  metrics                 = [tonumber(telnyx_ai_test_metric.resolved.id)]
  tags                    = ["booking"]
}
