## 1. Engine and generator

- [x] 1.1 Cursor paging (`page[after]` with `meta.cursors.after` or `meta.end_cursor`); page numbers win when both are offered; tests
- [x] 1.2 `singleton: true` with a fixed id put into every response; tests
- [x] 1.3 Envelope detection for `data` with `meta` or `links` siblings, or an object `data`
- [x] 1.4 Plain-text ephemeral responses as one `value` attribute; client accepts `text/plain`; tests
- [x] 1.5 Query lookups for data-source-only `read` overrides; tests
- [x] 1.6 Import of `id_from_path` settings takes every path parameter
- [x] 1.7 Codegen reports every failing override in one run

## 2. Wave 1 overrides

- [x] 2.1 44 resources with descriptions, sensitive secrets and tag sets (voice, numbers, porting, verify)
- [x] 2.2 26 read-only data sources and the lists of the wave 1 resources
- [x] 2.3 25 actions
- [x] 2.4 Ephemeral `telnyx_telephony_credential_token`
- [x] 2.5 Hook: porting order additional documents created as a list of one
- [x] 2.6 Coverage: LOA template download and TeXML queue list move to `out_of_scope`

## 3. Tests, docs and release notes

- [x] 3.1 Conformance test: every generated resource through create, empty plan, data source, import and destroy against a descriptor-driven fake API
- [x] 3.2 `make generate` docs; examples for the main wave 1 resources
- [x] 3.3 Changelog entry (`changie new`)
- [x] 3.4 `make check` passes
