# porting Specification

## Purpose
Manages porting phone numbers into Telnyx: porting orders, their documents and number ranges, and the steps of the porting workflow.

## Requirements

### Requirement: Porting orders
The provider SHALL manage `telnyx_porting_order` and its children `telnyx_porting_order_additional_document`, `telnyx_porting_order_associated_phone_number`, `telnyx_porting_order_phone_number_block` and `telnyx_porting_order_phone_number_extension`. Children SHALL be identified by `porting_order_id` and their own `id`, read from the order's lists, and import as `<porting_order_id>/<id>`. The end user's PIN SHALL be sensitive.

#### Scenario: Add a document
- **WHEN** a `telnyx_porting_order_additional_document` is created with `document_id` and `document_type`
- **THEN** the provider sends one `POST /porting_orders/{id}/additional_documents` with a list of that one document, and stores the one document the API returns

#### Scenario: Child removed outside Terraform
- **WHEN** a phone number block is no longer in its order's list
- **THEN** the next plan proposes to create it again

### Requirement: Porting workflow
The provider SHALL offer the porting workflow as actions: confirm (submit), activate, cancel, share, sending and verifying verification codes, initiating action requirements, updating activation jobs and setting phone number configurations. It SHALL offer data sources for porting phone numbers, reports, UK carriers, exception types, phone number configurations, and each order's requirements, action requirements, activation jobs, allowed FOC windows, verification codes and sub request.

#### Scenario: Submit an order
- **WHEN** `action "telnyx_porting_order_confirm"` is invoked for order `po1`
- **THEN** the provider sends `POST /porting_orders/po1/actions/confirm` once

#### Scenario: Share token
- **WHEN** `telnyx_porting_order_share` is invoked
- **THEN** the share is created, and the token is not available to the configuration because actions return no values

### Requirement: Porting order numbers
A porting order's `phone_numbers` SHALL be the set of numbers it was created with, compared with the numbers the order holds; a change SHALL force replacement.

#### Scenario: Read numbers
- **WHEN** the API returns the order's numbers as objects
- **THEN** `phone_numbers` holds their `phone_number` values and the next plan is empty
