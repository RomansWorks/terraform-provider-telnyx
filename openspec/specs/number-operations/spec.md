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
