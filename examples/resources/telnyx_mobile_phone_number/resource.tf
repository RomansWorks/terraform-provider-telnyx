# Manage a mobile number that is already on the account. The API cannot delete
# mobile numbers, so destroy only removes it from Terraform state.
resource "telnyx_mobile_phone_number" "main" {
  mobile_phone_number_id = "1293384261075731499"
  connection_id          = telnyx_credential_connection.office.id
  inbound_call_screening = "flag_calls"
  tags                   = ["mobile"]

  call_forwarding = {
    call_forwarding_enabled = true
    forwarding_type         = "on-failure"
    forwards_to             = "+13125550199"
  }
}
