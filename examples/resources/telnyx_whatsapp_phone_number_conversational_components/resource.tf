# Destroy leaves the components as they are.
resource "telnyx_whatsapp_phone_number_conversational_components" "support" {
  phone_number = telnyx_whatsapp_phone_number.support.phone_number
  ice_breakers = [
    "Where is my order?",
    "I want to return an item.",
  ]
  commands = [
    { command = "status", description = "Check the status of an order." },
    { command = "agent", description = "Talk to a person." },
  ]
}
