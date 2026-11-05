# Tasks

## 1. Engine, client and generator

- [x] 1.1 Record lists: `Op.Records`, `PagingPage` with `Op.SizeParam`, `decodeRecords`; tests
- [x] 1.2 Go names and list items names for types starting with a digit; tests
- [x] 1.3 String `const`, narrowed `allOf` properties, scalar `allOf`; tests
- [x] 1.4 Echoed request objects for free-form responses and maps; tests
- [x] 1.5 `Op.Whole` for free-form read responses; tests
- [x] 1.6 `snake` capital runs; `path_params` on lists; path parameter descriptions; tests
- [x] 1.7 Client: spec paths that repeat `/v2`; test
- [x] 1.8 Request-only fields when request and response types differ; records paging trusts `totalRecords`/`total_records`; `values` for any underivable read; tests

## 2. Wave 2 overrides

- [x] 2.1 10DLC: brands, campaigns, number assignments, shared campaigns, lookups and actions
- [x] 2.2 Messaging profiles, auto responses, alphanumeric sender ids, short codes, mobile numbers, hosted numbers and orders, toll-free verification
- [x] 2.3 Notifications and RCS
- [x] 2.4 WhatsApp numbers, templates, settings and actions
- [x] 2.5 Hooks: WhatsApp message template `waba_id`; inexplicit number order `count_requested` as a string
- [x] 2.6 Eight messaging operations the triage left unassigned: 10DLC enumerations, assignment tasks and assign action, RCS capability and test invite, URL domains, mobile number messaging settings
- [x] 2.7 Coverage: two multipart uploads and MNO metadata move to `out_of_scope`

## 3. Tests, docs and release notes

- [x] 3.1 Conformance test passes for every wave 2 resource; the fake answers `records` and bare-array lists
- [x] 3.2 `make generate` docs; examples for the main wave 2 resources
- [x] 3.3 Changelog entry
- [x] 3.4 Independent review; findings fixed
- [x] 3.5 `make check` passes; acceptance tests not run (no `TELNYX_API_KEY`)
