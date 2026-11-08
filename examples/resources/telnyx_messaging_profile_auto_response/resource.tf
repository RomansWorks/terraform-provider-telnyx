resource "telnyx_messaging_profile_auto_response" "stop" {
  messaging_profile_id = telnyx_messaging_profile.support.id
  op                   = "stop"
  keywords             = ["STOP", "UNSUBSCRIBE"]
  country_code         = "US"
  resp_text            = "You are unsubscribed from Acme alerts. Reply START to resubscribe."
}
