# number-operations Specification

## Purpose
Lookups and one-shot steps of the number workflow that no resource manages: number search, the phone numbers of number orders, sub number orders, regulatory requirements, bulk phone number jobs, port-out requests and infringement claims.

## Requirements

### Requirement: Number lookups
The provider SHALL offer read-only data sources for number search (`telnyx_available_phone_numbers`, `telnyx_available_phone_number_blocks`), the phone numbers of number orders (`telnyx_number_order_phone_number(s)`), sub number orders (`telnyx_sub_number_order(s)`), regulatory requirements (`telnyx_regulatory_requirements`) and infringement claims (`telnyx_infringement_claim`).

#### Scenario: Search numbers
- **WHEN** `data "telnyx_available_phone_numbers" "x" { filter_country_code = "US" }` is read
- **THEN** the provider sends `GET /available_phone_numbers` with `filter[country_code]=US`

### Requirement: Port-out lookups
The provider SHALL offer read-only data sources for port-out requests (`telnyx_portout`, `telnyx_portouts`), their supporting documents and rejection codes (`telnyx_portout_supporting_documents`, `telnyx_portout_rejections`) and port-out reports (`telnyx_portout_report`).

#### Scenario: Port-out
- **WHEN** `data "telnyx_portout" "p" { id = "po1" }` is read
- **THEN** the provider sends `GET /portouts/po1`

### Requirement: Number workflow actions
The provider SHALL offer actions that update the requirements of a number order's phone number or a sub number order, cancel a sub number order, start bulk phone number jobs (release, update, update emergency settings), authorize or reject a port-out and attach its supporting documents, schedule or cancel a requirement version, and contest an infringement claim. Each SHALL send its operation once when invoked. `telnyx_portout_status_update` SHALL take the new status as its `status` attribute.

#### Scenario: Reject a port-out
- **WHEN** `action "telnyx_portout_status_update" "x" { config { id = "po1" status = "rejected-pending" reason = "Wrong PIN" } }` is invoked
- **THEN** the provider sends `PATCH /portouts/po1/rejected-pending` once with the reason

#### Scenario: Release numbers in bulk
- **WHEN** `telnyx_phone_numbers_bulk_delete` is invoked with two numbers
- **THEN** the provider sends `POST /phone_numbers/jobs/delete_phone_numbers` once with both

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
