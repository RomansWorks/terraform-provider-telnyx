# Registering a brand with The Campaign Registry is billed and cannot be
# refunded.
resource "telnyx_10dlc_brand" "acme" {
  entity_type  = "PRIVATE_PROFIT"
  display_name = "Acme"
  company_name = "Acme Inc."
  ein          = "12-3456789"
  email        = "compliance@example.com"
  phone        = "+15551230000"
  street       = "1 Main St"
  city         = "Chicago"
  state        = "IL"
  postal_code  = "60601"
  country      = "US"
  website      = "https://example.com"
  vertical     = "TECHNOLOGY"
}
