# The API cannot delete RCS brands, so destroy only removes the brand from
# Terraform state.
resource "telnyx_rcs_brand" "acme" {
  display_name      = "Acme"
  legal_name        = "Acme Inc."
  legal_entity_type = "CORPORATION"
  organization_type = "PRIVATE_PROFIT"
  website_url       = "https://acme.example.com"
  profile_id        = telnyx_messaging_profile.support.id

  addresses = jsonencode({
    headquarters = {
      line_1              = "600 Congress Avenue"
      city                = "Austin"
      administrative_area = "TX"
      postal_code         = "78701"
      country_code        = "US"
    }
  })

  contacts = {
    brand = {
      contact_type = "BRAND"
      first_name   = "Jane"
      last_name    = "Doe"
      email        = "jane.doe@acme.example.com"
      phone_number = "+15125550100"
    }
  }

  identifiers = {
    ein = {
      identifier_type = "EIN"
      value           = "12-3456789"
    }
  }
}
