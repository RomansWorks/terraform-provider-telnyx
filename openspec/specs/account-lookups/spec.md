# account-lookups Specification

## Purpose
Read-only lookups of account objects that no product resource manages: connections of every type, organization users, Terms of Service, user tags, over-the-air SIM updates and compute actor types, plus removing a user and agreeing to terms.

## Requirements

### Requirement: Account lookups
The provider SHALL offer read-only data sources for connections of every type (`telnyx_connection`, `telnyx_connections`, `telnyx_connection_count`), organization users (`telnyx_organization_user`, `telnyx_organization_users`), Terms of Service (`telnyx_terms_of_service_info`, `telnyx_terms_of_service_status`, `telnyx_terms_of_service_agreement(s)`), user tags (`telnyx_user_tags`), over-the-air SIM updates (`telnyx_ota_update(s)`) and compute actor types (`telnyx_compute_actor`).

#### Scenario: Any connection
- **WHEN** `data "telnyx_connection" "c" { id = "c1" }` is read for an IP connection
- **THEN** the provider sends `GET /connections/c1` and exposes the fields common to all connection types

### Requirement: Account actions
The provider SHALL offer `telnyx_organization_user_remove`, `telnyx_terms_of_service_branded_calling_agree` and `telnyx_terms_of_service_number_reputation_agree`. Each SHALL send its operation once when invoked.

#### Scenario: Agree to terms
- **WHEN** `action "telnyx_terms_of_service_branded_calling_agree" "x" {}` is invoked
- **THEN** the provider sends `POST /terms_of_service/branded_calling/agree` once
