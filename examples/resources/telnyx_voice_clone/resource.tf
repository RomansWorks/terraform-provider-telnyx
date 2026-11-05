resource "telnyx_voice_design" "example" {
  name   = "Calm narrator"
  prompt = "A calm, warm voice with a slight British accent."
  text   = "Thanks for calling. How can I help you today?"
}

# A clone of a voice design.
resource "telnyx_voice_clone" "designed" {
  name            = "Calm narrator"
  voice_design_id = telnyx_voice_design.example.id
  language        = "en"
  gender          = "female"
}

# A clone of a recording of a real voice.
resource "telnyx_voice_clone" "recorded" {
  name       = "Founder"
  audio_path = "${path.module}/founder.wav"
  language   = "en"
  gender     = "male"
}
