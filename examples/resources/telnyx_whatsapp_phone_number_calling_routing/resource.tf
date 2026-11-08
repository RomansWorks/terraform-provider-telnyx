# Deliver inbound WhatsApp calls to a bring-your-own number to a Call Control
# application. Destroy leaves the routing as it is.
resource "telnyx_whatsapp_phone_number_calling_routing" "support" {
  phone_number  = "+13125550123"
  connection_id = telnyx_call_control_application.ivr.id
}
