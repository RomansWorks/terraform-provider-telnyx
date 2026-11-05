# The activation code of an eSIM, never stored in state.
ephemeral "telnyx_sim_card_activation_code" "phone" {
  sim_card_id = telnyx_sim_card.tracker.sim_card_id
}
