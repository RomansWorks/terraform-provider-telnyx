# A secret that TeXML templates use as {{crm_api_key}}. The value is
# write-only: bump value_wo_version to send a new value.
variable "crm_api_key" {
  type      = string
  sensitive = true
}

resource "telnyx_texml_secret" "crm" {
  name             = "crm_api_key"
  value_wo         = var.crm_api_key
  value_wo_version = 1
}
