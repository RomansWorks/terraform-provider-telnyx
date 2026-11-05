## ADDED Requirements

### Requirement: Balance and invoices
The provider SHALL offer read-only data sources for the account balance (`telnyx_balance`) and invoices (`telnyx_invoice`, `telnyx_invoices`).

#### Scenario: Balance
- **WHEN** `data "telnyx_balance" "b" {}` is read
- **THEN** the provider sends `GET /balance` and exposes the response fields

### Requirement: Charges
The provider SHALL offer `telnyx_charges_summary` and `telnyx_charges_breakdown`, which take the period as lookup attributes, required where the API requires them.

#### Scenario: Charges for a period
- **WHEN** `data "telnyx_charges_summary" "s" { start_date = "2026-10-01" end_date = "2026-11-01" }` is read
- **THEN** the provider sends both dates as query parameters

#### Scenario: Missing period
- **WHEN** `telnyx_charges_summary` has no `start_date`
- **THEN** validation fails before any request

### Requirement: Billing bundle lookups
The provider SHALL offer read-only data sources for the billing bundles Telnyx offers (`telnyx_bundle_pricing_billing_bundle(s)`) and those the account bought (`telnyx_bundle_pricing_user_bundle(s)`, `telnyx_bundle_pricing_unused_user_bundles`, `telnyx_bundle_pricing_user_bundle_resources`).

#### Scenario: Unused bundles
- **WHEN** `data "telnyx_bundle_pricing_unused_user_bundles" "u" {}` is read
- **THEN** the provider sends `GET /bundle_pricing/user_bundles/unused`

### Requirement: Billing bundle actions
The provider SHALL offer `telnyx_user_bundles_create`, which buys billing bundles, and `telnyx_user_bundle_deactivate`, which deactivates one. Each SHALL send its operation once when invoked.

#### Scenario: Deactivate a bundle
- **WHEN** `action "telnyx_user_bundle_deactivate" "x" { config { user_bundle_id = "ub1" } }` is invoked
- **THEN** the provider sends `DELETE /bundle_pricing/user_bundles/ub1` once
