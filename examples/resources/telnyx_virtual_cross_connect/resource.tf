variable "bgp_key" {
  type      = string
  sensitive = true
}

resource "telnyx_virtual_cross_connect" "aws" {
  name                     = "office-aws"
  network_id               = telnyx_network.office.id
  region_code              = "ashburn-va"
  cloud_provider           = "aws"
  cloud_provider_region    = "us-east-1"
  bgp_asn                  = 64512
  primary_cloud_account_id = "123456789012"
  primary_bgp_key          = var.bgp_key
}
