# Manage a SIM card that is already registered to the account. Destroy keeps
# the SIM card unless release_on_destroy is true.
resource "telnyx_sim_card" "tracker" {
  sim_card_id       = "6b14e151-8493-4fa1-8664-1cc4e6d14158"
  sim_card_group_id = telnyx_sim_card_group.fleet.id
  tags              = ["tracker"]
}
