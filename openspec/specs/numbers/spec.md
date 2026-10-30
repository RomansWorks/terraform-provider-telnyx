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
