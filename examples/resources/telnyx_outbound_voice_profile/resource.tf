resource "telnyx_outbound_voice_profile" "office" {
  name                      = "office"
  daily_spend_limit         = "50.00"
  daily_spend_limit_enabled = true
  whitelisted_destinations  = ["US", "CA"]

  call_recording = {
    call_recording_type   = "all"
    call_recording_format = "mp3"
  }
}
