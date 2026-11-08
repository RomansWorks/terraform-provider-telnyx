## ADDED Requirements

### Requirement: Phone number jobs and listings
The provider SHALL offer read-only data sources for bulk phone number jobs (`telnyx_phone_number_job(s)`), phone number block jobs (`telnyx_phone_number_block_job(s)`), the slim and voice-settings listings of the account's numbers (`telnyx_phone_numbers_slim`, `telnyx_phone_numbers_voice_settings`) and the regulatory requirements of given phone numbers (`telnyx_phone_number_regulatory_requirements`). The external PIN saved on an `update_phone_numbers` job SHALL be sensitive.

#### Scenario: Read a job
- **WHEN** `data "telnyx_phone_number_job" "j" { id = "j1" }` is read
- **THEN** the provider sends `GET /phone_numbers/jobs/j1` and `parameters.external_pin` is marked sensitive

#### Scenario: Slim listing with tags
- **WHEN** `data "telnyx_phone_numbers_slim" "s" { include_tags = true }` is read
- **THEN** the provider sends `GET /phone_numbers/slim` with `include_tags=true`

### Requirement: Phone number actions
The provider SHALL offer actions that enable emergency services on a phone number (`telnyx_phone_number_enable_emergency`), verify ownership of phone numbers (`telnyx_phone_number_verify_ownership`), assign a phone number to a regulatory bundle or remove it (`telnyx_phone_number_bundle_status_change`) and start a job that releases a phone number block (`telnyx_phone_number_block_delete`). Each SHALL send its operation once when invoked.

#### Scenario: Enable emergency services
- **WHEN** `action "telnyx_phone_number_enable_emergency" "e" { config { id = "pn1" emergency_enabled = true emergency_address_id = "addr1" } }` is invoked
- **THEN** the provider sends `POST /phone_numbers/pn1/actions/enable_emergency` once with both fields

#### Scenario: Release a number block
- **WHEN** `action "telnyx_phone_number_block_delete" "d" { config { phone_number_block_id = "blk1" } }` is invoked
- **THEN** the provider sends `POST /phone_number_blocks/jobs/delete_phone_number_block` once

### Requirement: Phone number bundle removal
`telnyx_phone_number_bundle_status_change` invoked without `bundle_id` SHALL remove the number from its regulatory bundle by sending `bundle_id` as null.

#### Scenario: Remove from bundle
- **WHEN** `action "telnyx_phone_number_bundle_status_change" "x" { config { id = "pn1" } }` is invoked
- **THEN** the provider sends `PATCH /phone_numbers/pn1/actions/bundle_status_change` once with `{"bundle_id":null}`
