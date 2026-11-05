resource "telnyx_room" "example" {
  unique_name       = "daily-standup"
  max_participants  = 10
  enable_recording  = false
  webhook_event_url = "https://example.com/webhooks/video"
}
