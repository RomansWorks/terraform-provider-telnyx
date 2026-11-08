# Destroy removes the voicemail settings from Terraform state only.
resource "telnyx_phone_number_voicemail" "example" {
  phone_number_id = "1293384261075731499"
  enabled         = true
  pin             = var.voicemail_pin
}

variable "voicemail_pin" {
  type      = string
  sensitive = true
}
