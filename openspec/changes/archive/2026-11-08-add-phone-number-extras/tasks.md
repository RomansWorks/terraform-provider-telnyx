## 1. Generator

- [x] 1.1 Attribute override key `nullable` for top-level action request fields: optional, sent as null when unset; an error elsewhere; codegen and engine tests

## 2. Overrides

- [x] 2.1 Job lookups: `telnyx_phone_number_jobs`, `telnyx_phone_number_job`, `telnyx_phone_number_block_jobs`, `telnyx_phone_number_block_job` (job `parameters.external_pin` sensitive)
- [x] 2.2 Listings: `telnyx_phone_numbers_slim`, `telnyx_phone_numbers_voice_settings`, `telnyx_phone_number_regulatory_requirements`
- [x] 2.3 Actions: `telnyx_phone_number_enable_emergency`, `telnyx_phone_number_verify_ownership`, `telnyx_phone_number_bundle_status_change` (`bundle_id` nullable), `telnyx_phone_number_block_delete`
- [x] 2.4 Provider tests for the actions and lookups against a stub API

## 3. Tests, docs and release notes

- [x] 3.1 `make generate` docs; examples for the data sources and actions
- [x] 3.2 Changelog entry (`changie new`)
- [x] 3.3 `make check` passes (acceptance tests not run: no API key)
