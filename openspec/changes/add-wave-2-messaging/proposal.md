# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 2 is messaging: messaging profiles and senders, hosted numbers, 10DLC registration, toll-free verification, notifications, RCS and WhatsApp. The coverage review assigned 26 resources, 42 data sources and 20 actions to it. Several APIs here have shapes the generator could not handle yet: 10DLC lists answer with `records` and page with `page` and `recordsPerPage`, type names start with a digit, camelCase fields with capital runs (`webhookURL`), maps of named entries in responses, constants without a type, and enums wrapped in `allOf`.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - Lists whose items are in `records` page with `page` and `recordsPerPage` (or `page_size`) and stop at a short page or `totalRecords`.
  - Type names starting with a digit get valid Go names, and list items attributes drop the digit segment (`telnyx_10dlc_brands` lists `brands`).
  - A string `const` without a type is a string with one allowed value; an `allOf` part that narrows a property refines it; an `allOf` of one scalar keeps its type and enum.
  - Response objects without fields, and maps without one entry shape, echo the request's fields, so import sees them.
  - Read responses that are maps without fixed fields become one JSON attribute `values`.
  - camelCase capital runs are one word (`webhook_url`); `path_params` also renames list parameters; path parameter descriptions no longer say "The phone number id." for values that are not ids.
  - The client sends spec paths that repeat `/v2` once (`/v2/mobile_phone_numbers` was sent as `/v2/v2/...`; this also fixes wave 1's mobile voice connections).
- **Wave 2 resources, data sources and actions**, generated from 67 new overrides. The list is in the `messaging` spec delta.
- **Hook:** WhatsApp message templates read `waba_id` from `whatsapp_business_account.id`.
- **Coverage.** Three operations move to `out_of_scope`: the hosted number order document upload and the WhatsApp profile photo upload (both multipart), and the 10DLC campaign MNO metadata (keyed by carrier id; the same data is in the use case qualification data source).

## Capabilities

### Modified Capabilities
- `codegen`: record paging, digit-leading names, consts and narrowed properties, echoed request objects, free-form read responses, versioned spec paths, capital runs in names.

### New Capabilities
- `messaging`: profiles and senders, short codes and mobile numbers, hosted numbers, 10DLC, toll-free verification, notifications, RCS and WhatsApp.

## Impact

- `internal/engine`: `Op.Records`, `Op.SizeParam`, `PagingPage`, `Op.Whole`.
- `internal/client`: version prefix handling.
- `internal/codegen`: `GoName`, `snake`, `norm` (allOf), `sources.child`, `listItems`, `listOp`, `DeriveList`, `deriveData`.
- `generator/overrides`: 67 new files. `internal/provider/generated`: regenerated.
- `internal/provider`: one hook and the conformance fake's list shapes.
- `catalog/coverage.yaml`: three entries move to `out_of_scope`.
- Docs, examples for the main resources, a changelog entry.
