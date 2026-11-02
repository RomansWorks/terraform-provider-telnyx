## ADDED Requirements

### Requirement: Actions
An override with `kind: action` SHALL produce a Terraform action whose configuration holds the operation's path parameters (Required) and request body fields. Invoking it SHALL send the operation once with the fields that are set, and an API error SHALL fail the action.

#### Scenario: Invoke an action
- **WHEN** `action "telnyx_ai_assistant_clone" "c" { config { assistant_id = "a1" } }` is invoked
- **THEN** the provider sends `POST /ai/assistants/a1/clone` once and stores nothing in state

#### Scenario: Action fails
- **WHEN** the API answers the action's request with an error
- **THEN** the action fails with the API's error message

### Requirement: Ephemeral resources
An override with `kind: ephemeral` SHALL produce an ephemeral resource. Its path parameters SHALL be Required and its response fields Computed and Sensitive. The value SHALL never be written to plan or state.

#### Scenario: Read a secret
- **WHEN** `ephemeral "telnyx_sim_card_activation_code" "x" { id = "sim1" }` is opened
- **THEN** its attributes hold the response of `GET /sim_cards/sim1/activation_code`, and neither plan nor state contains them

### Requirement: Write-only secrets
An attribute override with `write_only: true` SHALL produce a write-only, sensitive attribute named `<name>_wo` and an optional `<name>_wo_version`. Create SHALL send the write-only value. Update SHALL send it only when `<name>_wo_version` changed. The value SHALL never be stored in state.

#### Scenario: Create with a secret
- **WHEN** a resource is created with `password_wo = "s3cret"`
- **THEN** the create request carries `password`, and state holds no `password_wo` value

#### Scenario: Rotate a secret
- **WHEN** `password_wo` changes but `password_wo_version` does not
- **THEN** the plan is empty
- **WHEN** `password_wo_version` changes as well
- **THEN** the update request carries the new `password`

### Requirement: Release on destroy
An override with `release_on_destroy: true` SHALL add an optional `release_on_destroy` attribute, false by default and never sent to the API. Destroy SHALL call the delete operation only when it is true in state. Otherwise destroy SHALL remove the object from state and warn that it still exists in Telnyx.

#### Scenario: Keep by default
- **WHEN** a phone number with no `release_on_destroy` is destroyed
- **THEN** no delete request is sent, the number leaves state, and a warning says it is still in Telnyx

#### Scenario: Release
- **WHEN** a phone number with `release_on_destroy = true` in state is destroyed
- **THEN** the delete operation is sent
