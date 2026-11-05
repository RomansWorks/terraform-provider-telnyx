# Enable a SIM card once it is in a group. Needs Terraform 1.14 or later.
resource "terraform_data" "enable_tracker" {
  input = telnyx_sim_card.tracker.sim_card_group_id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.telnyx_sim_card_enable.tracker]
    }
  }
}

action "telnyx_sim_card_enable" "tracker" {
  config {
    sim_card_id = telnyx_sim_card.tracker.sim_card_id
  }
}
