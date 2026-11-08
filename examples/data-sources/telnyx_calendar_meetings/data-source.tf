data "telnyx_calendar_meetings" "next_week" {
  calendar_connection_id = telnyx_calendar_connection.team.calendar_connection_id
}

output "meetings_the_bot_joins" {
  value = [for m in data.telnyx_calendar_meetings.next_week.calendar_meetings : m.id if m.bot_will_join]
}
