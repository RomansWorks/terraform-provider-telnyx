resource "telnyx_ai_conversation_insight" "sentiment" {
  name         = "Customer sentiment"
  instructions = "Rate the customer's sentiment from 1 (angry) to 5 (delighted)."
}

resource "telnyx_ai_conversation_insight_group_assignment" "sentiment" {
  group_id   = telnyx_ai_conversation_insight_group.support_quality.id
  insight_id = telnyx_ai_conversation_insight.sentiment.id
}
