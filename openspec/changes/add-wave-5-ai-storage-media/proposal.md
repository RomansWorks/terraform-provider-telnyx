# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 5 is AI, storage and media: the AI platform around assistants (knowledge collections, memory, missions, insights, the test framework and the AI gateway), Telnyx Storage beyond buckets (KV, SQL, CloudFS, migrations), media files, video rooms, voice clones and designs, and compute function log exports. The coverage review assigned 40 resources, 44 data sources and 14 actions to it. Several have shapes the generator could not handle yet: secrets shown only in the create response, a raw-bytes key-value API, lists in `results`, a required `Idempotency-Key`, `If-Match` versioning, and request bodies that are a union per provider.

This change covers the Telnyx product areas AI, storage and media. It covers no upstream range: it builds on the current pin.

## What Changes

- **Client.** Every POST, PUT, PATCH and DELETE carries a random `Idempotency-Key`, the same on each retry of that request. Hooks can add request headers.
- **Generator and engine.**
  - `create_only: true` keeps a field that only the create response carries, such as a key shown once, in state.
  - Request bodies of raw bytes (`application/octet-stream`) are allowed when a `Send` hook supplies them; a read that answers with raw bytes or plain text is one `value` attribute.
  - Lists with items in `results` and their count in `count` page with `page` and `page_size`, like wave 2's `records` lists.
  - Lookup query parameters of data sources can be booleans and integers, not only strings.
  - `merge_variants: true` turns a union of object variants with compatible fields (one request shape per provider) into one object.
  - Hooks: `Send` replaces a create or update request; `Headers` adds headers to update and delete.
  - Generated files for types ending in `_test` get a `_resource` suffix, so Go does not take them for tests.
- **Wave 5 resources, data sources and actions**, generated from overrides. The list is in `tasks.md`.
- **Hand-written behaviour** (hooks): KV key values, bucket SSL certificate upload, voice clones from a recording, and AI gateway versioning.
- **Coverage.** These move to `out_of_scope`, each with its reason: mission knowledge bases, MCP servers and tools (the spec gives no schema), assistant tool attachments (use `tool_ids` on `telnyx_ai_assistant`), collection settings read (read with the collection), assistant TeXML (a bare string), test metric bulk create, and conversation insight deletion (runtime data). `telnyx_ai_collection_settings` becomes part of `telnyx_ai_collection`, and the upload of a voice recording becomes a way to create `telnyx_voice_clone` instead of an action.

## Capabilities

### New Capabilities
- `ai-platform`: knowledge collections, memory namespaces, missions, conversation insights, the AI test framework and the AI gateway.
- `storage`: KV namespaces and keys, SQL databases, CloudFS filesystems, migrations and bucket SSL certificates.
- `media`: media files and video rooms.
- `speech`: voice clones, voice designs and pronunciation dictionaries.
- `compute`: compute function log exports and function lookups.

### Modified Capabilities
- `ai-assistants`: assistant tags, canary deployments, tests, versions, integration secrets and the assistant actions.
- `codegen`: create-only response fields, raw bodies, `results` lists, typed query lookups, merged unions, `Send` and `Headers` hooks.
- `provider-core`: idempotency keys.

## Impact

- `internal/client`: `Idempotency-Key`, `WithHeader`, `Accept` also takes `application/octet-stream`.
- `internal/engine`: `Attr.CreateOnly`, `Op.Raw`, `Op.Results`, `Hooks.Send`, `Hooks.Headers`, typed query lookups, raw reads.
- `internal/codegen`: override keys `create_only` and `merge_variants`; raw bodies; `results` lists; typed query lookups; file names.
- `generator/overrides`: 62 new override files. `internal/provider/generated`: regenerated.
- `internal/provider`: hooks for KV keys, bucket SSL certificates, voice clones and the AI gateway; tests against stub APIs for each.
- `catalog/coverage.yaml`: wave 5 entries reviewed; the entries above move to `out_of_scope`.
- Docs and examples for the new resources; a changelog entry.
