resource "telnyx_sim_card_group" "fleet" {
  name = "fleet"
  data_limit = {
    amount = "2048"
    unit   = "MB"
  }
  private_wireless_gateway_id = telnyx_private_wireless_gateway.office.id
  wireless_blocklist_id       = telnyx_wireless_blocklist.no_roaming.id
}
