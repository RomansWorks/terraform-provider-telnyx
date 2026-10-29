resource "telnyx_ai_assistant" "front_desk" {
  name         = "Front desk"
  instructions = "You answer questions about our opening hours and book appointments."
  greeting     = "Hello, how can I help?"
  tool_ids     = [telnyx_ai_tool.order_status.id]

  voice_settings = {
    voice = "Telnyx.NaturalHD.astra"
  }
}
