resource "telnyx_porting_order" "example" {
  phone_numbers = ["+13125550123", "+13125550124"]

  end_user = {
    admin = {
      entity_name          = "Acme Inc."
      auth_person_name     = "Jane Doe"
      account_number       = "123456789"
      billing_phone_number = "+13125550123"
    }
    location = {
      street_address      = "600 Congress Avenue"
      locality            = "Austin"
      administrative_area = "TX"
      postal_code         = "78701"
      country_code        = "US"
    }
  }

  documents = {
    loa     = telnyx_document.loa.id
    invoice = telnyx_document.invoice.id
  }

  phone_number_configuration = {
    connection_id = telnyx_credential_connection.office.id
  }
}
