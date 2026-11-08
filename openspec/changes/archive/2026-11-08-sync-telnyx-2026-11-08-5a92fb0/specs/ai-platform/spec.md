## ADDED Requirements

### Requirement: Meeting bot calendar connections
The provider SHALL manage `telnyx_calendar_connection`, identified by `calendar_connection_id`. Because a connection is created by browser OAuth, create SHALL adopt an existing connection with `PATCH /calendar_connections/{id}`. Destroy SHALL remove it from state and warn, unless `release_on_destroy` is true in state, when it SHALL disconnect it with `DELETE /calendar_connections/{id}`.

#### Scenario: Adopt a connection
- **WHEN** a configuration sets `calendar_connection_id` and `auto_join_policy = "all_meetings"`
- **THEN** create sends `PATCH /calendar_connections/{id}` with the policy, and the next plan is empty

#### Scenario: Destroy keeps the calendar connected
- **WHEN** a connection without `release_on_destroy` is destroyed
- **THEN** no request is sent, the connection leaves state, and a warning says it is still connected

#### Scenario: Disconnect on destroy
- **WHEN** a connection with `release_on_destroy = true` in state is destroyed
- **THEN** the provider sends `DELETE /calendar_connections/{id}`

### Requirement: Calendar meetings
`data "telnyx_calendar_meetings"` SHALL list a calendar connection's tracked meetings, and the action `telnyx_calendar_meeting_auto_join` SHALL set one meeting's auto-join override. Starting calendar OAuth and its callback SHALL be `out_of_scope`.

#### Scenario: Opt one meeting out
- **WHEN** the action runs with `bot_will_join = false`
- **THEN** the provider sends `PATCH /calendar_connections/{id}/meetings/{meeting_id}` with `{"bot_will_join": false}`
