# Zones cannot be created or deleted; destroy only removes this from state.
resource "telnyx_voice_channel_zone" "example" {
  channel_zone_id = var.channel_zone_id
  channels        = 10
}

variable "channel_zone_id" {
  type = string
}
