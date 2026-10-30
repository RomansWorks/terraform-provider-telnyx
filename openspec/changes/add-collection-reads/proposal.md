# Proposal

## Why

The coverage review (`2026-10-30-triage-api-coverage`) set the finish line at 157 resources and 297 data sources. The engine cannot yet build several of those shapes:

- **130 list data sources** (`GET` collections), such as `telnyx_billing_groups`.
- **About 167 read-only data sources** that belong to no resource.
- **15 resources with no item `GET`**, which can only be read back from their list operation (integration secrets, spend limits, voice clones, porting order children and others).
- **4 attachment resources** that can only be read from their parent object (assistant tags, tool attachments, email IP pool IPs, insight group assignments).
- **3 resources whose settings are changed through separate endpoints**: SIM card group, managed account and enterprise reputation.

Every rollout wave needs these features, so they come first. This is part 1 of step 2. Part 2 covers actions, ephemeral resources, write-only secrets and `release_on_destroy`.

This change covers no upstream range: it builds on the current pin, `team-telnyx/openapi@3d1ee2b`.

## What Changes

- **List data sources.** An override can name a resource's list operation, or define a data source on its own. The generator derives:
  - filter attributes from the query parameters
  - a computed list of items from the list response
  - the pagination style

  The engine fetches every page, up to an optional `max_items`.
- **Standalone data sources.** An override file can describe a data source that has no resource: a single object or a list.
- **Read from a list.** A resource with no item `GET` reads itself by finding its id in the list operation. When the id is missing, the resource is removed from state, the same as a 404.
- **Read from the parent.** A resource can read itself from a field of its parent object. The field is either an array of objects matched by a key, or an array of plain values.
- **Extra update endpoints.** A resource can name further operations that each set some of its attributes. Optionally, each can name an operation that clears them. The engine calls only the operations whose attributes changed.

No existing resource or data source changes behaviour.

## Capabilities

### Modified Capabilities
- `codegen`: list and standalone data sources, read from list or parent, and extra update endpoints.

## Impact

- `internal/engine`: new descriptor fields, a `List` descriptor, a list data source, pagination, and read-from-collection.
- `internal/codegen`: new override keys, query parameter and pagination derivation, and goldens.
- `generator/overrides`: the 7 pilots gain their list data sources (`telnyx_ai_assistants` and others), so this change is user-visible and gets a changelog entry.
