resource "telnyx_public_internet_gateway" "example" {
  name        = "office-internet"
  network_id  = telnyx_network.example.id
  region_code = "ashburn-va"
}
