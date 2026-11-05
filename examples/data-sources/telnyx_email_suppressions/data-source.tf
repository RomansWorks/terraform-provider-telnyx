data "telnyx_email_suppressions" "bounces" {
  filter_reason = "hard_bounce"
}
