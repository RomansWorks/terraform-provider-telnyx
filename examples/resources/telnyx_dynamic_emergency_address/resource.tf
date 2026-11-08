# Addresses cannot be changed, so every change replaces the address.
resource "telnyx_dynamic_emergency_address" "example" {
  house_number        = "600"
  street_name         = "Congress"
  street_suffix       = "Ave"
  extended_address    = "Suite 200"
  locality            = "Austin"
  administrative_area = "TX"
  postal_code         = "78701"
  country_code        = "US"
}
