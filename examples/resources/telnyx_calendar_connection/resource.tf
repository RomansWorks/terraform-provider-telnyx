# Connect the calendar first (POST /calendar_connections and the browser
# OAuth consent), then manage its meeting bot defaults here.
resource "telnyx_calendar_connection" "team" {
  calendar_connection_id = "<calendar_connection_id>"

  auto_join_policy    = "all_meetings"
  bot_name            = "Notetaker"
  default_recording   = true
  default_webhook_url = "https://example.com/meeting-events"

  default_recording_export = {
    bucket = "meeting-recordings"
    prefix = "calendar/"
  }

  # Disconnect the calendar when this resource is destroyed.
  release_on_destroy = true
}
