resource "telnyx_porting_order_phone_number_block" "example" {
  porting_order_id = telnyx_porting_order.example.id
  phone_number_range = {
    start_at = "+13125550100"
    end_at   = "+13125550199"
  }
  activation_ranges = [
    { start_at = "+13125550100", end_at = "+13125550109" },
  ]
}
