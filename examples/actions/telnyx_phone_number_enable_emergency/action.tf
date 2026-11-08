# Turn on emergency calling for a number once its address exists. Needs
# Terraform 1.14 or later.
resource "terraform_data" "emergency" {
  input = telnyx_address.office.id

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.telnyx_phone_number_enable_emergency.main]
    }
  }
}

action "telnyx_phone_number_enable_emergency" "main" {
  config {
    id                   = telnyx_phone_number.main.id
    emergency_enabled    = true
    emergency_address_id = telnyx_address.office.id
  }
}
