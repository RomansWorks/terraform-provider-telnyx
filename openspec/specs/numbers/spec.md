# numbers Specification

## Purpose
Manages buying phone numbers and the settings of numbers already on the account.

## Requirements

### Requirement: Number order
The provider SHALL manage `telnyx_number_order` with `POST /number_orders` and `GET /number_orders/{number_order_id}`. After create it SHALL wait until `status` is `success`, failing on `failure` or when `timeouts.create` (default 10 minutes) passes. Changes other than `customer_reference` and `regulatory_requirements` SHALL force replacement; those two SHALL update with `PATCH /number_orders/{number_order_id}`.

#### Scenario: Order completes
- **WHEN** a user orders a number and the order moves from `pending` to `success`
- **THEN** apply succeeds and the ordered numbers are in state

#### Scenario: Order fails
- **WHEN** the order reaches `failure`
- **THEN** apply fails with the order's error details

### Requirement: Number order destroy
Destroying a number order SHALL remove it from state and warn that the ordered numbers stay on the account until released.

#### Scenario: Destroy
- **WHEN** a number order is destroyed
- **THEN** it leaves state with a warning naming the phone numbers

### Requirement: Phone number voice settings
The provider SHALL manage `telnyx_phone_number_voice_settings` for an existing number, identified by `phone_number_id` (forces replacement), using `GET` and `PATCH /phone_numbers/{id}/voice`. Create SHALL apply the configured settings; destroy SHALL leave the settings unchanged in Telnyx and remove them from state.

#### Scenario: Adopt settings
- **WHEN** a user applies voice settings for a number
- **THEN** the configured settings are PATCHed and the next plan is empty

#### Scenario: Destroy
- **WHEN** the resource is destroyed
- **THEN** no API call is made and the settings stay as they were

### Requirement: Phone number
The provider SHALL manage `telnyx_phone_number` for a number already on the account, identified by `phone_number_id` (forces replacement). Create SHALL apply the configured settings with `PATCH /phone_numbers/{id}`. Destroy SHALL keep the number on the account and warn, unless `release_on_destroy = true`, in which case it SHALL release the number with `DELETE /phone_numbers/{id}`. `tags` SHALL be a set.

#### Scenario: Adopt a number
- **WHEN** a user applies `telnyx_phone_number` with `phone_number_id` and `tags`
- **THEN** the tags are PATCHed to the number and the next plan is empty

#### Scenario: Keep on destroy
- **WHEN** a phone number without `release_on_destroy` is destroyed
- **THEN** no delete request is sent and a warning says the number is still on the account

#### Scenario: Release on destroy
- **WHEN** a phone number with `release_on_destroy = true` is destroyed
- **THEN** the provider sends `DELETE /phone_numbers/{id}`

### Requirement: Phone number settings
The provider SHALL manage `telnyx_phone_number_messaging_settings` and `telnyx_phone_number_voicemail` for an existing number, identified by `phone_number_id` (forces replacement). Create SHALL apply the configured settings; destroy SHALL leave them unchanged in Telnyx and remove them from state. The voicemail `pin` SHALL be sensitive.

#### Scenario: Destroy settings
- **WHEN** messaging settings are destroyed
- **THEN** no API call is made and the settings stay as they were

### Requirement: Number orders and reservations
The provider SHALL manage `telnyx_advanced_number_order`, `telnyx_inexplicit_number_order`, `telnyx_number_block_order` and `telnyx_number_reservation`. They have no update or delete: a change SHALL force replacement, and destroy SHALL remove them from state with a warning. `telnyx_number_reservation_extend` SHALL extend a reservation.

#### Scenario: Destroy an order
- **WHEN** a number block order is destroyed
- **THEN** no API call is made, it leaves state, and a warning says it still exists in Telnyx

### Requirement: Addresses, documents and requirements
The provider SHALL manage `telnyx_address`, `telnyx_user_address`, `telnyx_dynamic_emergency_address`, `telnyx_dynamic_emergency_endpoint`, `telnyx_document`, `telnyx_requirement_group`, `telnyx_verified_number` and `telnyx_billing_group`. Objects the API cannot change SHALL be replaced when their configuration changes; objects the API cannot delete SHALL leave state with a warning on destroy. A document's base64 `file` SHALL be sensitive.

#### Scenario: Change an address
- **WHEN** the street of a `telnyx_address` changes
- **THEN** the plan replaces the address

#### Scenario: Verified number id
- **WHEN** a `telnyx_verified_number` is created for `+13125550100`
- **THEN** the create request carries `phone_number`, the resource is identified by it, and import takes the phone number

### Requirement: Number workflow actions and lookups
The provider SHALL offer the wave 1 number actions (accepting address suggestions, setting requirement groups on orders, submitting a requirement group for approval, fulfilling sub number order requirements, bulk messaging profile updates and verifying a verified number) and data sources (document links, messaging settings of all numbers, bulk update status and sub number order requirements).

#### Scenario: Submit a requirement group
- **WHEN** `action "telnyx_requirement_group_submit_for_approval" "x" { config { id = "rg1" } }` is invoked
- **THEN** the provider sends `POST /requirement_groups/rg1/submit_for_approval` once
