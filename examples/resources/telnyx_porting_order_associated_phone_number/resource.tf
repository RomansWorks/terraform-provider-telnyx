# In a partial port, keep the rest of the range with the current carrier.
resource "telnyx_porting_order_associated_phone_number" "example" {
  porting_order_id = telnyx_porting_order.example.id
  action           = "keep"
  phone_number_range = {
    start_at = "+13125550125"
    end_at   = "+13125550129"
  }
}
