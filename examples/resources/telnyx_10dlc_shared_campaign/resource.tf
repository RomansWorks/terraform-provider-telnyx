# Manage the webhooks of a campaign a partner shared with Telnyx. The partner
# owns the campaign, so destroy only removes it from Terraform state.
resource "telnyx_10dlc_shared_campaign" "partner" {
  campaign_id          = "CX1Y2Z3"
  webhook_url          = "https://example.com/webhooks/10dlc"
  webhook_failover_url = "https://backup.example.com/webhooks/10dlc"
}
