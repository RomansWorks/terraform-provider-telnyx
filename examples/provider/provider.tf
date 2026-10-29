terraform {
  required_providers {
    telnyx = {
      source = "romansworks/telnyx"
    }
  }
}

# The API key can also come from the TELNYX_API_KEY environment variable.
provider "telnyx" {
  api_key = var.telnyx_api_key
}

variable "telnyx_api_key" {
  type      = string
  sensitive = true
}
