# media Specification

## Purpose
How the provider manages media files and video rooms.

## Requirements

### Requirement: Media files
The provider SHALL manage `telnyx_media`, identified and imported by `media_name`, created from `media_url` with the JSON form of `POST /media`. Its list data source SHALL be `telnyx_media_files`.

#### Scenario: Replace the file
- **WHEN** `media_url` changes
- **THEN** the provider sends `PUT /media/{media_name}` and the next plan is empty

### Requirement: Video rooms
The provider SHALL manage `telnyx_room`. A room's sessions SHALL not be attributes: they are its call history.

#### Scenario: Room lifecycle
- **WHEN** a room is created, its `max_participants` changed with PATCH, destroyed and imported
- **THEN** each step uses the matching operation and every plan after apply is empty

### Requirement: Room client tokens
The provider SHALL offer the ephemeral `telnyx_room_client_token`, minted with `POST /rooms/{room_id}/actions/generate_join_client_token`, with optional `token_ttl_secs` and `refresh_token_ttl_secs`. `token`, `refresh_token` and their expiry times SHALL never be stored in plan or state. Refreshing a token SHALL stay `out_of_scope`: a client does it with its refresh token.

#### Scenario: Token for a guest
- **WHEN** a configuration opens the ephemeral for a room and passes `token` to a write-only argument
- **THEN** a token is minted on each run and does not appear in plan or state
