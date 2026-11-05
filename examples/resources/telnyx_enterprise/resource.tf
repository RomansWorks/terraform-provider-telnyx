resource "telnyx_enterprise" "acme" {
  legal_name                    = "Acme Inc."
  doing_business_as             = "Acme"
  organization_type             = "commercial"
  organization_legal_type       = "corporation"
  country_code                  = "US"
  website                       = "https://acme.example.com"
  fein                          = "12-3456789"
  industry                      = "business"
  number_of_employees           = "51-200"
  jurisdiction_of_incorporation = "DE"

  organization_contact = {
    first_name   = "Ada"
    last_name    = "Lovelace"
    email        = "ada@acme.example.com"
    job_title    = "CTO"
    phone_number = "+13125550100"
  }
  billing_contact = {
    first_name   = "Grace"
    last_name    = "Hopper"
    email        = "billing@acme.example.com"
    phone_number = "+13125550101"
  }
  organization_physical_address = {
    country             = "US"
    administrative_area = "IL"
    city                = "Chicago"
    postal_code         = "60601"
    street_address      = "1 Main St"
  }
  billing_address = {
    country             = "US"
    administrative_area = "IL"
    city                = "Chicago"
    postal_code         = "60601"
    street_address      = "1 Main St"
  }
}
