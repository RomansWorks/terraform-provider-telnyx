resource "telnyx_pay_connector" "example" {
  name = "acme-processor"
  type = "generic"
  mode = "test"

  config_wo = jsonencode({
    endpoint_url = "https://processor.example.com/pay"
    auth_type    = "bearer"
    token        = var.processor_token
  })
  config_wo_version = 1
}

variable "processor_token" {
  type      = string
  sensitive = true
}
