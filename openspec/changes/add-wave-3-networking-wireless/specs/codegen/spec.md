## ADDED Requirements

### Requirement: Values only create returns
An attribute override MAY set `create_only: true` for a top-level response field that no request takes. The attribute SHALL be computed and sensitive. When the read response schema lacks the field, the generator SHALL take it from the create response schema, and SHALL fail when neither lists it. The engine SHALL fill the value from the create response and SHALL keep the value in state on every read. Setting `create_only` on a field a request takes SHALL be a generation error.

#### Scenario: Token only create returns
- **WHEN** `telnyx_tunnel` is created and `POST /tunnels` answers with `tunnel_token`
- **THEN** `tunnel_token` is in state, shown as sensitive
- **WHEN** the tunnel refreshes and `GET /tunnels/{id}` has no `tunnel_token`
- **THEN** the value in state is unchanged and the plan is empty

#### Scenario: Read answers null
- **WHEN** a WireGuard peer refreshes and `GET /wireguard_peers/{id}` answers `"private_key": null`
- **THEN** the private key from create stays in state

### Requirement: Read-only request objects
A request object whose fields are all `readOnly` SHALL not be configurable: the attribute SHALL be computed from the response. An object with at least one writable field SHALL stay configurable.

#### Scenario: SIM card status
- **WHEN** `PATCH /sim_cards/{id}` lists `status` with only read-only fields
- **THEN** `telnyx_sim_card.status` is computed and cannot be set in configuration
