# Assign a short code on the account to a messaging profile. Short codes cannot
# be deleted, so destroy only removes the settings from Terraform state.
resource "telnyx_short_code_settings" "alerts" {
  short_code_id        = "a0f9e7b8-1c2d-4e5f-8a9b-0c1d2e3f4a5b"
  messaging_profile_id = telnyx_messaging_profile.support.id
  tags                 = ["alerts"]
}
