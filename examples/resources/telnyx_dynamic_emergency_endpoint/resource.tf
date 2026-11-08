resource "telnyx_dynamic_emergency_endpoint" "example" {
  dynamic_emergency_address_id = telnyx_dynamic_emergency_address.example.id
  callback_number              = "+15125550100"
  caller_name                  = "Acme Austin office"
}
