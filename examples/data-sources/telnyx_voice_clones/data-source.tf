data "telnyx_voice_clones" "active" {
  filter_provider = "telnyx"
}

output "voice_ids" {
  value = { for v in data.telnyx_voice_clones.active.voice_clones : v.name => v.provider_voice_id }
}
