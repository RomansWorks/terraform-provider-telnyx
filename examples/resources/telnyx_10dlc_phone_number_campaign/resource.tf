resource "telnyx_10dlc_phone_number_campaign" "main" {
  phone_number = "+15551230000"
  campaign_id  = telnyx_10dlc_campaign.alerts.id
}
