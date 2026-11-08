resource "telnyx_requirement_group" "example" {
  country_code      = "DE"
  phone_number_type = "local"
  action            = "ordering"

  regulatory_requirements = [
    {
      requirement_id = var.address_requirement_id
      field_value    = telnyx_address.example.id
    },
  ]
}

variable "address_requirement_id" {
  type = string
}
