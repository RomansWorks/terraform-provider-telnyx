## ADDED Requirements

### Requirement: Nullable action fields
An attribute override with `nullable: true` on a top-level field of an action request SHALL make the attribute optional; invoking the action with the attribute unset SHALL send the field as JSON `null`, and with a value SHALL send the value. `nullable` on any other attribute SHALL be a generation error.

#### Scenario: Remove a number from its bundle
- **WHEN** `action "telnyx_phone_number_bundle_status_change" "x" { config { id = "pn1" } }` is invoked
- **THEN** the provider sends `PATCH /phone_numbers/pn1/actions/bundle_status_change` once with `{"bundle_id":null}`

#### Scenario: Nullable outside an action
- **WHEN** a data source override sets `nullable: true` on an attribute
- **THEN** generation fails naming the attribute
