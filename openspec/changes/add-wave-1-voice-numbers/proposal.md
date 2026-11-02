# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 1 is voice and numbers: the objects most Telnyx accounts start with. The coverage review assigned 44 resources, 26 read-only data sources, 25 actions and 1 ephemeral resource to it. Several of them have shapes the generator could not handle yet: lists paged by cursor, settings that exist once per account, a token returned as plain text, a lookup that takes query parameters, and responses with `meta` next to `data`.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - Lists paged with `page[after]` follow `meta.cursors.after` or `meta.end_cursor`. Page numbers still win when an operation offers both.
  - `singleton: true` makes a resource for settings that exist once per account. Its id is fixed.
  - A response with `data` and envelope siblings such as `meta` reads the object from `data`.
  - An ephemeral resource whose operation answers in plain text exposes the text as one attribute.
  - A data-source-only override with `read` takes the operation's scalar query parameters as lookup attributes.
  - Settings objects identified by a path parameter (`id_from_path`) import with every path parameter, parents first.
- **Wave 1 resources, data sources, actions and the ephemeral token**, generated from overrides. The list is in `tasks.md`.
- **`telnyx_phone_number`** adopts a number already on the account. Destroy keeps it unless `release_on_destroy = true`, as decided for numbers and SIM cards.
- **Porting.** Porting orders, their documents, number blocks, extensions and associated numbers become resources, and the order workflow (confirm, activate, cancel, verification codes) becomes actions.
- **Coverage.** Two operations move to `out_of_scope`: the LOA template download (a PDF) and the TeXML queue list (a different list shape; single queues are still readable).
- **A conformance test** runs every generated resource through create, an empty plan, its data source, import and destroy against a fake API driven by its descriptor.

## Capabilities

### Modified Capabilities
- `codegen`: cursor paging, once-per-account settings, envelopes with metadata, plain-text ephemeral values, data source query lookups, import of settings objects.
- `numbers`: phone numbers and their settings, number orders other than the plain order, reservations and verified numbers.
- `voice`: connections, applications, queues, channel settings and the other wave 1 voice objects.

### New Capabilities
- `porting`: porting orders and their workflow.

## Impact

- `internal/engine`: `PagingCursor`, `Resource.FixedID`, `Attr.Query`, `Op.Text`, `Resource.Path` for hooks.
- `internal/client`: requests also accept `text/plain`.
- `internal/codegen`: the override key `singleton`, envelope detection, cursor paging, query lookups, plain-text responses, import ids.
- `generator/overrides`: 97 new override files. `internal/provider/generated`: regenerated.
- `internal/provider`: one hook (porting order additional documents are created as a list of one) and the conformance test.
- `catalog/coverage.yaml`: two entries move to `out_of_scope`.
- Docs and examples for the new resources; a changelog entry.
