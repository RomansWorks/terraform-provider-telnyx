# Creating it sends a verification code to the number.
resource "telnyx_verified_number" "example" {
  phone_number        = "+13125550123"
  verification_method = "sms"
}
