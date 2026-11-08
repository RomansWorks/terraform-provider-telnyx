# Destroy leaves the setting as it is.
resource "telnyx_whatsapp_phone_number_calling_settings" "support" {
  phone_number = telnyx_whatsapp_phone_number.support.phone_number
  enabled      = true
}
