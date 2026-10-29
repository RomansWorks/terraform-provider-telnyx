# Tasks

## 1. Engine

- [x] 1.1 Implement value conversion between Terraform values and JSON for every type in design decision 2, verified by table-driven unit tests
- [x] 1.2 Implement create, read (404 removal), update (PATCH diff, PUT full), delete, missing-delete warning and import, verified by engine unit tests against httptest
- [x] 1.3 Implement hook points (before create, custom create, after read, wait) and `timeouts`, verified by unit tests

## 2. Generator

- [ ] 2.1 Define the override file format and load it with validation, verified by unit tests rejecting unknown operations and attributes
- [ ] 2.2 Derive schemas and descriptors from the spec per design decisions 2 to 5, verified by golden-file tests on the fixture spec
- [ ] 2.3 Emit resources, data sources, docs templates and example skeletons; wire `make generate` and verify a second run leaves no diff

## 3. Pilot resources

Each task: override, hooks if any, unit tests against recorded responses, acceptance test with import step, example under `examples/`, changie entry, coverage entries reviewed.

- [ ] 3.1 `telnyx_ai_mcp_server` and data source
- [ ] 3.2 `telnyx_outbound_voice_profile` and data source
- [ ] 3.3 `telnyx_ai_tool` and data source, including the `tool_definition` mapping hook
- [ ] 3.4 `telnyx_ai_assistant` and data source
- [ ] 3.5 `telnyx_ai_avatar` and data source, including the upload hook
- [ ] 3.6 `telnyx_phone_number_voice_settings` and data source
- [ ] 3.7 `telnyx_number_order` with the wait hook, acceptance test gated by `TELNYX_ACC_BILLABLE=1`

## 4. Integration

- [ ] 4.1 Add sweepers for every pilot resource and verify `go test ./internal/... -sweep=all` lists them
- [ ] 4.2 Run `make check` and the acceptance tests that need no billing, and record results in the PR
