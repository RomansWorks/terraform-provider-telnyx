# Submit the DIR for vetting once it is created. Needs Terraform 1.14 or
# later.
resource "terraform_data" "submit" {
  input = telnyx_display_identity_record.acme.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.telnyx_display_identity_record_submit.acme]
    }
  }
}

action "telnyx_display_identity_record_submit" "acme" {
  config {
    dir_id = telnyx_display_identity_record.acme.id
  }
}
