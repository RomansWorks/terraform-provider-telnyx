resource "telnyx_ai_assistant_test" "example" {
  name         = "Refund request"
  destination  = "+13125550100"
  instructions = "You bought headphones last week and want a refund."
  test_suite   = "billing"
  rubric = [
    { name = "Policy", criteria = "The assistant explains the 30-day refund policy." },
    { name = "Tone", criteria = "The assistant stays polite throughout." },
  ]
}
