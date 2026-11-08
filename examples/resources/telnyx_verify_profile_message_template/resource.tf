# The API cannot delete templates, so destroy only removes it from state.
resource "telnyx_verify_profile_message_template" "example" {
  text = "Your {{app_name}} verification code is: {{code}}."
}
