# Block voice once it costs more than 50 USD in a day.
resource "telnyx_spend_limit" "voice_daily" {
  product = "voice"
  period  = "daily"
  amount  = 50
  reason  = "Cap test traffic"
}
