# Proposal

## Why

The coverage review classified 11 phone number operations as data sources or actions, but no wave built them: `make check` passes because coverage records the bucket, not whether an override exists. They are the bulk job lookups, the slim and voice listings of the account's numbers, regulatory requirements per phone number, and the enable-emergency, verify-ownership, bundle-status and block-delete actions. With them, every operation classified `data_source` or `action` is implemented.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator.** A new attribute override key, `nullable: true`, for top-level fields of an action request: the attribute is optional, and leaving it unset sends JSON `null`. Telnyx removes a number from its bundle only when `bundle_id` is `null`, and a required Terraform attribute cannot be null.
- **Data sources:** `telnyx_phone_number_jobs`, `telnyx_phone_number_job`, `telnyx_phone_number_block_jobs`, `telnyx_phone_number_block_job`, `telnyx_phone_numbers_slim`, `telnyx_phone_numbers_voice_settings`, `telnyx_phone_number_regulatory_requirements`.
- **Actions:** `telnyx_phone_number_enable_emergency`, `telnyx_phone_number_verify_ownership`, `telnyx_phone_number_bundle_status_change`, `telnyx_phone_number_block_delete`.
- The saved `external_pin` of an `update_phone_numbers` job is sensitive.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `codegen`: nullable action fields.
- `number-operations`: phone number job, listing and regulatory requirement lookups, and four phone number actions.

## Impact

- `internal/codegen` and `internal/engine`: the `nullable` key and the null it sends, with tests.
- `generator/overrides`: 11 new override files. `internal/provider/generated`: regenerated.
- Docs and examples; a changelog entry.
- No change to existing resources. Terraform actions return no values, so `telnyx_phone_number_verify_ownership` reports only success or an API error, not which numbers were found.
