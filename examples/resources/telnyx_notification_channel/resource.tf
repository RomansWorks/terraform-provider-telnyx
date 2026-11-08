resource "telnyx_notification_profile" "ops" {
  name = "Operations"
}

resource "telnyx_notification_channel" "ops_email" {
  notification_profile_id = telnyx_notification_profile.ops.id
  channel_type_id         = "email"
  channel_destination     = "ops@example.com"
}
