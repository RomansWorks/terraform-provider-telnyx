## ADDED Requirements

### Requirement: Room client tokens
The provider SHALL offer the ephemeral `telnyx_room_client_token`, minted with `POST /rooms/{room_id}/actions/generate_join_client_token`, with optional `token_ttl_secs` and `refresh_token_ttl_secs`. `token`, `refresh_token` and their expiry times SHALL never be stored in plan or state. Refreshing a token SHALL stay `out_of_scope`: a client does it with its refresh token.

#### Scenario: Token for a guest
- **WHEN** a configuration opens the ephemeral for a room and passes `token` to a write-only argument
- **THEN** a token is minted on each run and does not appear in plan or state
