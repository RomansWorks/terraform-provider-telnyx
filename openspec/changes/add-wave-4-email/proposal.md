# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 4 is email: the sending and receiving domains, inboxes, dedicated IPs and their pools, templates and suppressions that a Telnyx email setup consists of. The coverage review assigned 11 resources, 12 data sources and 7 actions to it. Two shapes were new to the generator: deletes that need flags to remove an object in use (a verified domain, a pool with IPs), and a list that answers in one of two page shapes.

This change covers the Telnyx email product area and no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - `force_destroy: [<param>, ...]` adds a local `force_destroy` attribute. When it is true in state, destroy sends each named boolean query parameter of the delete operation as `true`.
  - A list response whose schema is a `oneOf` or `anyOf` of page shapes reads its items from the first variant with a `data` array.
- **Wave 4 resources, data sources and actions**, generated from overrides. The list is in `tasks.md`.
- **Read hooks** for `telnyx_email_inbox` (username from the address) and `telnyx_email_domain` (`inbound_enabled` from `inbound.enabled`), so both detect drift and import.
- **Coverage.** Three operations move to `out_of_scope`: `PATCH /email/dedicated_ips/{id}` (one imperative change per request; pool membership and warm-up have their own resource and actions), and the suppression CSV import and its status (`POST /email_blocks/import`, `GET /email_blocks/import/{id}`).
- **The conformance test's fake API** answers reads from a parent object's array field, and keys a member by the id its create request names.

## Capabilities

### Modified Capabilities
- `codegen`: `force_destroy`, union list responses.

### New Capabilities
- `email`: email domains and their webhooks, inboxes and sender filters, dedicated IPs and IP pools, templates, suppressions and unsubscribe groups.

## Impact

- `internal/engine`: `Resource.ForceDestroy`, `ForceDestroyAttr`; delete sends the query parameters.
- `internal/codegen`: the override key `force_destroy`; `listItems` falls back to union variants.
- `generator/overrides`: 19 new override files. `internal/provider/generated`: regenerated.
- `internal/provider`: two read hooks; the conformance fake learns parent-field reads.
- `catalog/coverage.yaml`: three entries move to `out_of_scope`.
- Docs and examples for the new resources; a changelog entry.
