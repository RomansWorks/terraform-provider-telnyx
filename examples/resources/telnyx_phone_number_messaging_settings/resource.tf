# Destroy removes the settings from Terraform state only.
resource "telnyx_phone_number_messaging_settings" "example" {
  phone_number_id      = "1293384261075731499"
  messaging_profile_id = telnyx_messaging_profile.support.id
}
