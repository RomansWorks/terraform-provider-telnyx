resource "telnyx_texml_queue" "example" {
  account_sid   = var.account_sid
  friendly_name = "support"
  max_size      = 50
}

variable "account_sid" {
  type = string
}
