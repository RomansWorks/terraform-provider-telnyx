# The account's inbound voice channels in the US zone. There is one per
# account.
resource "telnyx_inbound_voice_channel_settings" "this" {
  channels = 10
}
