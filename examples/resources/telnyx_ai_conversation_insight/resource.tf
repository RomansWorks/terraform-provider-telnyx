resource "telnyx_ai_conversation_insight" "example" {
  name         = "Customer sentiment"
  instructions = "Rate the customer's sentiment from 1 (angry) to 5 (delighted)."
}

resource "telnyx_ai_conversation_insight_group" "example" {
  name = "Support quality"
}

resource "telnyx_ai_conversation_insight_group_assignment" "example" {
  group_id   = telnyx_ai_conversation_insight_group.example.id
  insight_id = telnyx_ai_conversation_insight.example.id
}
