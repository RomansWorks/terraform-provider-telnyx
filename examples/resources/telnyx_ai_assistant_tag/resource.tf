resource "telnyx_ai_assistant_tag" "example" {
  assistant_id = telnyx_ai_assistant.example.id
  tag          = "production"
}
