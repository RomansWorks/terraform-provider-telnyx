# Keep the meeting bot out of one meeting. Needs Terraform 1.14 or later.
action "telnyx_calendar_meeting_auto_join" "skip_board_meeting" {
  config {
    calendar_connection_id = telnyx_calendar_connection.team.calendar_connection_id
    meeting_id             = "<meeting_id>"
    bot_will_join          = false
  }
}
