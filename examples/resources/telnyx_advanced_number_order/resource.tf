# Orders cannot be cancelled through the API.
resource "telnyx_advanced_number_order" "example" {
  country_code      = "US"
  phone_number_type = "local"
  area_code         = "312"
  quantity          = 2
  features          = ["sms", "voice"]
}
