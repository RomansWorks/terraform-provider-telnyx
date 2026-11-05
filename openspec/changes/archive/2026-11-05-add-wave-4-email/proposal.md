# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 4 is email: the sending and receiving domains, inboxes, dedicated IPs and their pools, templates and suppressions that a Telnyx email setup consists of. The coverage review assigned 11 resources, 12 data sources and 7 actions to it; this change ships the 11 resources, 20 data sources (10 lists, 9 single objects and domain health, plus DNS records) and 6 actions, and moves the CSV suppression import to `out_of_scope`. Several shapes were new to the generator: deletes that need flags to remove an object in use (a verified domain, a pool with IPs), a list that answers in one of two page shapes, objects the API soft-deletes, and email addresses the API stores in lowercase.

This change covers the Telnyx email product area and no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - `force_destroy: [<param>, ...]` adds a local `force_destroy` attribute. When it is true in state, destroy sends each named boolean query parameter of the delete operation as `true`.
  - A list response whose schema is a `oneOf` or `anyOf` of page shapes reads its items from the first variant with a `data` array.
  - `lowercase: true` on an attribute rejects values (or set and list elements) that are not lowercase or have surrounding spaces, because the API stores them normalised and every plan would show a change.
  - Soft-deleted objects (a status that says deleted) are reported as not found by a read hook, which removes them from state.
  - A resource without an update operation stores a change of its local attributes (`force_destroy`, `release_on_destroy`) without a request, instead of failing.
- **Wave 4 resources, data sources and actions**, generated from overrides. The list is in `tasks.md`.
- **Read hooks** for `telnyx_email_inbox` (username from the address) and `telnyx_email_domain` (`inbound_enabled` from `inbound.enabled`), so both detect drift and import; and for `telnyx_email_suppression` (status `removed`) and `telnyx_email_dedicated_ip` (status `released`), which the API soft-deletes.
- **Coverage.** Three operations move to `out_of_scope`: `PATCH /email/dedicated_ips/{id}` (one imperative change per request; pool membership and warm-up have their own resource and actions), and the suppression CSV import and its status (`POST /email_blocks/import`, `GET /email_blocks/import/{id}`).
- **The conformance test's fake API** answers reads from a parent object's array field, and keys a member by the id its create request names.

## Capabilities

### Modified Capabilities
- `codegen`: `force_destroy`, union list responses, `lowercase`, soft deletes, local-only changes without an update.

### New Capabilities
- `email`: email domains and their webhooks, inboxes and sender filters, dedicated IPs and IP pools, templates, suppressions and unsubscribe groups.

## Impact

- `internal/engine`: `Resource.ForceDestroy`, `ForceDestroyAttr`, `Attr.Lowercase`; delete sends the query parameters; update without an update operation.
- `internal/codegen`: the override keys `force_destroy` and `attributes.<f>.lowercase`; `listItems` falls back to union variants.
- `generator/overrides`: 19 new override files. `internal/provider/generated`: regenerated.
- `internal/provider`: four read hooks; the conformance fake learns parent-field reads.
- `catalog/coverage.yaml`: three entries move to `out_of_scope`.
- Docs and examples for the new resources; a changelog entry.
