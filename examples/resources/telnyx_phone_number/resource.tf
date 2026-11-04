# Manage a number that is already on the account. Destroy keeps the number
# unless release_on_destroy is true.
resource "telnyx_phone_number" "main" {
  phone_number_id  = "1293384261075731499"
  connection_id    = telnyx_credential_connection.office.id
  billing_group_id = telnyx_billing_group.office.id
  tags             = ["office"]
}
