# The webhooks that receive WhatsApp signup events. There is one per account;
# destroy leaves its values as they are.
resource "telnyx_whatsapp_user_data" "this" {
  webhook_url          = "https://example.com/webhooks/whatsapp-signup"
  webhook_failover_url = "https://backup.example.com/webhooks/whatsapp-signup"
}
