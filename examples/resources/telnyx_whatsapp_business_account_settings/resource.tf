# Manage the settings of an existing WhatsApp Business Account. Destroy leaves
# them as they are.
data "telnyx_whatsapp_business_account" "main" {
  id = var.whatsapp_business_account_id
}

resource "telnyx_whatsapp_business_account_settings" "main" {
  whatsapp_business_account_id = data.telnyx_whatsapp_business_account.main.id
  name                         = "Acme Support"
  timezone                     = "America/Chicago"
  webhook_enabled              = true
  webhook_url                  = "https://example.com/webhooks/whatsapp"
  webhook_failover_url         = "https://backup.example.com/webhooks/whatsapp"
}

variable "whatsapp_business_account_id" {
  type = string
}
