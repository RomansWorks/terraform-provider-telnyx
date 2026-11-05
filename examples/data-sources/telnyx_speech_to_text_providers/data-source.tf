# Deepgram models that AI assistants can use. `provider_filter` is sent as
# the API's `provider` query parameter, a name Terraform reserves.
data "telnyx_speech_to_text_providers" "deepgram" {
  provider_filter = "deepgram"
  service_type    = "ai_assistant"
}
