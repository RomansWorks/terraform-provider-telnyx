## 1. Client and engine

- [x] 1.1 `Idempotency-Key` on every POST, PUT, PATCH and DELETE, reused across retries; `client.WithHeader`; `Accept` takes plain text and octet-stream
- [x] 1.2 `Op.Raw` with the `Send` hook; text reads become one `value` attribute
- [x] 1.3 `Op.Results`: lists with items in `results` and the count in `count`, paged by `page` and `page_size`
- [x] 1.4 Typed lookup query parameters (booleans and integers)
- [x] 1.5 `Headers` hook on update and delete
- [x] 1.6 A `Delete` hook also destroys objects without a delete operation (merged with wave 3's hook)
- [x] 1.7 Unit tests for each

## 2. Generator

- [x] 2.1 `create_only` searches every 2xx create response (on top of wave 3's implementation)
- [x] 2.2 Raw request bodies and text responses
- [x] 2.3 `results` lists and typed query lookups
- [x] 2.4 `merge_variants` for unions of object variants per provider, after wave 3's boolean-flag merge
- [x] 2.5 Generated file names for types ending in `_test`
- [x] 2.6 Fixtures and tests in `internal/codegen/wave5_test.go`; the codegen spec delta

## 3. Resources, data sources and actions

- [x] 3.1 AI assistants: tags, canary deployments, tests, versions, integration secrets, and the clone, import, restore and version actions
- [x] 3.2 AI platform: knowledge collections and sources, memory namespaces and settings, missions, conversation insights, groups and assignments, the AI test framework (agents, metrics, scenarios, schedules), AI gateway token groups and keys
- [x] 3.3 Storage: KV namespaces and keys, SQL databases with the query action, CloudFS filesystems with the meta token rotation, migrations and sources, bucket SSL certificates
- [x] 3.4 Media files, video rooms, voice clones, voice designs, pronunciation dictionaries, compute function log exports and lookups
- [x] 3.5 Agent Gateway: the gateway, APIs, guardrails, keys, MCP servers, teams, policies, policy attachments, key and team tool permissions, list data sources and models
- [x] 3.6 Real descriptions everywhere; secrets sensitive, `create_only` or write-only; unordered lists as sets
- [x] 3.7 Hooks: KV key raw values, bucket SSL certificate upload, voice clones from a design or a recording, AI gateway `If-Match` and versions, collection retrieval settings, Agent Gateway readiness, tool permission destroy
- [x] 3.8 Coverage: wave 5 entries reviewed; mission knowledge bases, MCP servers and tools, tool attachments, collection settings read, assistant TeXML, metric bulk create and insight deletion moved to `out_of_scope` with notes

## 4. Tests, docs and release notes

- [x] 4.1 Stub API tests for the resources the conformance fake cannot play, each in `conformanceSkip` with its reason
- [x] 4.2 Examples for the new resources, data sources and actions; storage buckets guide using the AWS provider with Telnyx's S3 endpoints
- [x] 4.3 `make generate` (descriptors and docs)
- [x] 4.4 Changelog entry
- [x] 4.5 `go test ./...`, `go vet`, `apisync check`, `openspec validate --strict`
- [x] 4.6 Independent review and fixes
