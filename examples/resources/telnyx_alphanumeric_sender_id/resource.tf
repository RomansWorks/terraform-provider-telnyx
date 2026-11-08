# The API cannot change a sender ID, so every change replaces it.
resource "telnyx_alphanumeric_sender_id" "acme" {
  alphanumeric_sender_id = "ACME"
  messaging_profile_id   = telnyx_messaging_profile.support.id
  us_long_code_fallback  = "+13125550100"
}
