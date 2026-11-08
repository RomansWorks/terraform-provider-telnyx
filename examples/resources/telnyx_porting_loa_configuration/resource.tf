resource "telnyx_porting_loa_configuration" "example" {
  name         = "Acme LOA"
  company_name = "Acme Inc."

  address = {
    street_address = "600 Congress Avenue"
    city           = "Austin"
    state          = "TX"
    zip_code       = "78701"
    country_code   = "US"
  }

  contact = {
    email        = "porting@acme.example.com"
    phone_number = "+15125550100"
  }

  logo = {
    document_id = telnyx_document.logo.id
  }
}
