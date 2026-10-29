# Buys the listed numbers. Telnyx cannot cancel an order: destroying this
# resource only removes it from Terraform state, and the numbers stay in
# your account until you release them.
resource "telnyx_number_order" "main_line" {
  customer_reference = "main-line"
  phone_numbers = [
    { phone_number = "+13125550100" },
  ]

  timeouts = {
    create = "20m"
  }
}
