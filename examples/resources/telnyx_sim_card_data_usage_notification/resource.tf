resource "telnyx_sim_card_data_usage_notification" "example" {
  sim_card_id = telnyx_sim_card.example.id
  threshold = {
    amount = "5"
    unit   = "GB"
  }
}
