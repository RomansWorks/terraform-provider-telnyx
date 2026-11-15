# Tasks

## 1. Sync

- [x] 1.1 Pin `bf2eb8e`, rebuild the catalog (fine-tuning entries removed)
- [x] 1.2 Review the 27 changed operations: `make generate` adds two enum values and changes descriptions only

## 2. Coverage gaps

- [x] 2.1 Codegen and engine: `read: none` (unreadable resources), with tests
- [x] 2.2 Codegen and engine: request body arguments for ephemeral resources, with tests
- [x] 2.3 `telnyx_texml_secret` override, coverage entry and example
- [x] 2.4 `telnyx_room_client_token` and `telnyx_storage_object_presigned_url` overrides, coverage entries and examples
- [x] 2.5 Conformance test expects import to be refused for unreadable resources

## 3. Check

- [x] 3.1 Changelog entries
- [x] 3.2 `make check` passes
- [x] 3.3 Acceptance tests: not run, `TELNYX_API_KEY` is not set
