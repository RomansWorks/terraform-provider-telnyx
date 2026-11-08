# Proposal

## Why

The coverage review classified every Telnyx operation, but 76 data sources and actions were not assigned to any of the six product waves. Wave 7 builds them, so that every operation classified `data_source` or `action` is implemented. They are mostly account-wide lookups (balance, invoices, charges, connections of every type), reference data (regions, coverage, pricing, speech-to-text providers) and number workflow steps (sub number orders, port-outs, bulk phone number jobs).

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator.** A list filter named like a Terraform meta-argument (`provider`, `count`) takes the suffix `_filter`, and is still sent under its own name. Before, it was a generation error.
- **54 data sources and 18 actions**, generated from overrides. The list is in `tasks.md`.
- **Left to other waves.** `telnyx_bulk_sim_card_action(s)` and `telnyx_wireless_regions` are built by wave 3, and `telnyx_telephony_credential_token_material` by wave 6.
- **A skipped lookup parameter.** `telnyx_charges_breakdown` leaves out `format` (`csv` is not JSON), with the `skip` key wave 6 added. `telnyx_organization_user` takes `include_groups` and `telnyx_requirement` takes `version`, with the boolean and integer lookups wave 5 added.

## Capabilities

### New Capabilities
- `billing`: balance, invoices, charges and billing bundles.
- `reference-data`: regions, coverage, pricing, requirement types and other lists Telnyx maintains.
- `number-operations`: number search, number order and sub number order lookups and steps, regulatory requirements, bulk phone number jobs, port-outs and infringement claims.
- `account-lookups`: connections of every type, organization users, Terms of Service, user tags, OTA updates and compute actor types.
- `ai-integrations`: the integrations AI assistants can use and the account's connections to them.

### Modified Capabilities
- `codegen`: list filters with reserved names.

## Impact

- `internal/codegen`: the reserved-name rule for list filters, with a test.
- `generator/overrides`: 72 new override files. `internal/provider/generated`: regenerated.
- Docs and examples; a changelog entry.
- No change to existing resources.
