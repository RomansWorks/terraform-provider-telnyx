# Manages the voice settings of an existing phone number. Destroying this
# resource leaves the settings as they are.
resource "telnyx_phone_number_voice_settings" "main_line" {
  phone_number_id        = "1293384261075731499"
  inbound_call_screening = "flag_calls"

  call_forwarding = {
    call_forwarding_enabled = true
    forwards_to             = "+13125550100"
    forwarding_type         = "on-failure"
  }
}
