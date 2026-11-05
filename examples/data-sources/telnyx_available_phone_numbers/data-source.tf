# Up to 10 local US numbers with SMS that start with +1312.
data "telnyx_available_phone_numbers" "chicago" {
  filter_country_code             = "US"
  filter_phone_number_type        = "local"
  filter_phone_number_starts_with = "312"
  filter_features                 = ["sms"]
  max_items                       = 10
}
