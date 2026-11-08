# Destroy removes the settings from Terraform state only.
resource "telnyx_external_connection_phone_number_settings" "example" {
  external_connection_id = telnyx_external_connection.example.id
  phone_number_id        = "1293384261075731499"
  location_id            = var.location_id
}

variable "location_id" {
  type = string
}
