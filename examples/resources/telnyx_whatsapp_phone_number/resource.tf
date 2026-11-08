# Creating the number sends a verification code to it. Complete registration
# with the telnyx_whatsapp_phone_number_verify action.
resource "telnyx_whatsapp_phone_number" "support" {
  whatsapp_business_account_id = var.whatsapp_business_account_id
  phone_number                 = "+13125550123"
  display_name                 = "Acme Support"
  verification_method          = "sms"
}
