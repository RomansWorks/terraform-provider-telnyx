## MODIFIED Requirements

### Requirement: Operation coverage
`catalog/coverage.yaml` SHALL have exactly one entry for every operation in the pinned spec, keyed by `METHOD /path`, with a bucket of `resource`, `data_source`, `action`, `ephemeral` or `out_of_scope`. `resource` and `ephemeral` entries SHALL name a Terraform type, and so SHALL reviewed `data_source` and `action` entries. `out_of_scope` entries SHALL give a reason. `make check` SHALL fail on missing, stale or invalid entries.

#### Scenario: New upstream operation
- **WHEN** a sync adds an operation
- **THEN** `make catalog` adds an entry with a suggested bucket and `reviewed: false`, and leaves existing entries untouched

#### Scenario: Removed upstream operation
- **WHEN** a sync removes an operation
- **THEN** `make catalog` removes its entry and the API diff lists it

#### Scenario: Release readiness
- **WHEN** `apisync check -require-reviewed` runs while any entry has `reviewed: false`
- **THEN** it fails

#### Scenario: Reviewed entry without a name
- **WHEN** an entry with bucket `data_source` or `action` has `reviewed: true` and no `terraform` name
- **THEN** `apisync check` reports it as invalid

## ADDED Requirements

### Requirement: Aliased paths
When two operations share a method and operationId and their paths differ only by `/` versus `_` or a leading `/v2`, the review SHALL keep one canonical operation: the one with `x-latency-category`, then the one not under `/v2/`, then the one with fewer segments. The others SHALL be `out_of_scope` with the note `alias of <canonical key>`.

#### Scenario: Aliased path
- **WHEN** `GET /oauth_clients` and `GET /oauth/clients` share the operationId `ListOAuthClients`, and only `/oauth/clients` has `x-latency-category`
- **THEN** `GET /oauth_clients` is `out_of_scope` with the note `alias of GET /oauth/clients`, and only one `telnyx_oauth_client` exists

### Requirement: Resource lifecycle entries
Operations that create, read, update or delete one configuration object SHALL be `resource` entries naming that object's type, even when the object has no item `GET`. The collection `GET` that lists those objects SHALL be a `data_source` with the plural name. When the API offers both `PUT` and `PATCH` for the update, the unused one SHALL be `out_of_scope` with the note `alternative update`.

#### Scenario: Configuration object without an item read
- **WHEN** an object has `POST /integration_secrets` and `DELETE /integration_secrets/{id}` but no item `GET`
- **THEN** both are `resource` entries for `telnyx_integration_secret`, and `GET /integration_secrets` is the data source `telnyx_integration_secrets`

### Requirement: Runtime and reporting operations
Operations on runtime traffic and its data SHALL be `out_of_scope` with the reason as the note. Runtime traffic covers calls, conferences, messages, faxes, chats, model inference, live sessions, recordings, mailbox content, test and mission runs, and OAuth flows. Reports, usage, metrics, logs, events and downloads SHALL be `out_of_scope` with the note `reporting`. Lookups and previews that change nothing SHALL also be `out_of_scope`.

#### Scenario: Runtime object with a create and delete
- **WHEN** a path such as `/faxes/{id}` has create, read and delete operations but describes traffic rather than configuration
- **THEN** its entries are `out_of_scope` with a runtime reason, not a resource

### Requirement: Secrets, lookups and one-shot operations
An operation that returns a secret value, such as a token, an activation code or a config containing keys, SHALL be `ephemeral` so the value never enters state. Any other `GET` SHALL be a `data_source`. Any other mutation, such as submitting, enabling, verifying, rotating or a bulk job, SHALL be an `action`.

#### Scenario: Token endpoint
- **WHEN** the review reaches `POST /telephony_credentials/{id}/token`
- **THEN** it is the ephemeral resource `telnyx_telephony_credential_token`, not an action or a data source
