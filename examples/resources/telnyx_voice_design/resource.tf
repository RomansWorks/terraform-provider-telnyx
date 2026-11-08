resource "telnyx_voice_design" "narrator" {
  name     = "Warm narrator"
  prompt   = "A warm, calm female voice with a slight British accent, speaking slowly."
  text     = "Welcome to Acme. Your call is important to us."
  language = "English"
}
