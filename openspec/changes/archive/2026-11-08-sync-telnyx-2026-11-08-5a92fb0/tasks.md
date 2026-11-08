# Tasks

## 1. Classify and implement

- [x] 1.1 Move `telnyx_mobile_phone_number` and `telnyx_mobile_voice_connection` overrides to the paths without `/v2`
- [x] 1.2 Add `telnyx_calendar_connection` (adopt, `release_on_destroy`), its data sources, `telnyx_calendar_meetings` and `telnyx_calendar_meeting_auto_join`
- [x] 1.3 Classify the OAuth, stored payment and meeting recording operations `out_of_scope`
- [x] 1.4 Review the 49 changed operations: `make generate` changes only descriptions and adds porting `remediation`
- [x] 1.5 Examples and changelog entries

## 2. Sync workflow

- [x] 2.1 `sync.yml` sets up Terraform, runs `make generate` and commits generated descriptors, docs and examples

## 3. Check

- [x] 3.1 `make api-check-release` passes (no unreviewed operations)
- [x] 3.2 `make check` passes (unit and conformance tests)
- [x] 3.3 Acceptance tests: not run, `TELNYX_API_KEY` is not set and they were excluded on purpose
