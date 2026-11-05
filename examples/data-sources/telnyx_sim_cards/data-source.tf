# Every SIM card tagged "tracker".
data "telnyx_sim_cards" "trackers" {
  filter_tags = ["tracker"]
}
