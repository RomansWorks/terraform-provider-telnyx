resource "telnyx_porting_order_phone_number_extension" "example" {
  porting_order_id        = telnyx_porting_order.example.id
  porting_phone_number_id = var.porting_phone_number_id
  extension_range = {
    start_at = 100
    end_at   = 199
  }
  activation_ranges = [
    { start_at = 100, end_at = 149 },
  ]
}

variable "porting_phone_number_id" {
  type = string
}
