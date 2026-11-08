resource "telnyx_number_reservation" "example" {
  phone_numbers = [
    { phone_number = "+13125550123" },
  ]
}
