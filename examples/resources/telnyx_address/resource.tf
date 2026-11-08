# Addresses cannot be changed, so every change replaces the address.
resource "telnyx_address" "example" {
  business_name       = "Acme Inc."
  first_name          = "Jane"
  last_name           = "Doe"
  street_address      = "600 Congress Avenue"
  extended_address    = "Suite 200"
  locality            = "Austin"
  administrative_area = "TX"
  postal_code         = "78701"
  country_code        = "US"
  address_book        = true
}
