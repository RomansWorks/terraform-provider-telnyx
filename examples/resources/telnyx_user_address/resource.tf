# The API cannot change or delete user addresses.
resource "telnyx_user_address" "example" {
  business_name       = "Acme Inc."
  first_name          = "Jane"
  last_name           = "Doe"
  street_address      = "600 Congress Avenue"
  locality            = "Austin"
  administrative_area = "TX"
  postal_code         = "78701"
  country_code        = "US"
}
