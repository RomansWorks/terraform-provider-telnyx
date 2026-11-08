# Orders cannot be cancelled through the API.
resource "telnyx_number_block_order" "example" {
  starting_number = "+13125550100"
  range           = 10
  connection_id   = telnyx_credential_connection.office.id
}
