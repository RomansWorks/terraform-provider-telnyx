## MODIFIED Requirements

### Requirement: Ephemeral resources
An override with `kind: ephemeral` SHALL produce an ephemeral resource. Its path parameters SHALL be Required. The fields of its operation's request body SHALL be arguments, Optional unless the spec requires them, and neither Computed nor Sensitive; a request field that is an object SHALL be a generator error. Its response fields SHALL be Computed and Sensitive. Opening it SHALL send a JSON body of the arguments the configuration sets whenever the operation takes a body. The value SHALL never be written to plan or state.

#### Scenario: Read a secret
- **WHEN** `ephemeral "telnyx_sim_card_activation_code" "x" { id = "sim1" }` is opened
- **THEN** its attributes hold the response of `GET /sim_cards/sim1/activation_code`, and neither plan nor state contains them

#### Scenario: Arguments in the request body
- **WHEN** `ephemeral "telnyx_room_client_token" "t" { room_id = "r1", token_ttl_secs = 60 }` is opened
- **THEN** the provider sends `POST /rooms/r1/actions/generate_join_client_token` with `{"token_ttl_secs": 60}`, and an argument left unset is not sent

## ADDED Requirements

### Requirement: Objects the API cannot read
An override MAY set `read: none` for an object the API can create but never read back. Its attributes SHALL come from the create request and response, it SHALL have no data source, and `update`, `updates`, `read_from` and `list` SHALL be generator errors with it. Refresh SHALL keep the state without a request, and import SHALL fail with "Import not supported".

#### Scenario: Refresh sends nothing
- **WHEN** a `telnyx_texml_secret` created with one `POST /texml/secrets` is refreshed and planned again
- **THEN** no request is sent and the plan is empty

#### Scenario: Import refused
- **WHEN** `terraform import telnyx_texml_secret.x my_secret` runs
- **THEN** it fails with "Import not supported" and nothing is added to state
