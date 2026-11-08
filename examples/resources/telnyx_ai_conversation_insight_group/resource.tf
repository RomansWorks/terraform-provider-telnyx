resource "telnyx_ai_conversation_insight_group" "support_quality" {
  name        = "Support quality"
  description = "Insights that grade support conversations."
  webhook     = "https://example.com/webhooks/insights"
}
