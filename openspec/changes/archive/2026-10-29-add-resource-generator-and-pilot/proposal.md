# Proposal

## Why

The provider has foundations but no resources. Writing 168+ resources by hand would be slow and would drift from the API, so resources must come from the pinned OpenAPI spec through a generator. A pilot set that exercises every awkward Telnyx pattern proves the generator before the rollout by product area.

Covers upstream `team-telnyx/openapi@3d1ee2b` (the current pin) for these product areas: Assistants, Avatars, Outbound Voice Profiles, Phone Number Configurations and Phone Number Orders.

## What Changes

- A generator (`make generate`) that turns the pinned spec plus one small override file per resource into Terraform Plugin Framework schemas, request and response mapping, and test and doc scaffolding. A single shared engine runs create, read, update, delete and import for every generated resource.
- Seven pilot resources, each chosen for a pattern the rollout will meet again:

| Resource | Pattern it proves |
|---|---|
| `telnyx_ai_assistant` | Very large nested schema; update with `POST` |
| `telnyx_ai_tool` | Variant attributes (webhook, handoff, …); response shape differs from request |
| `telnyx_ai_mcp_server` | Full-replacement update with `PUT` |
| `telnyx_ai_avatar` | Image upload (multipart); no delete in the API |
| `telnyx_outbound_voice_profile` | Responses wrapped in `data`; classic CRUD |
| `telnyx_phone_number_voice_settings` | Settings of an object created elsewhere; adopt and release |
| `telnyx_number_order` | Asynchronous: waits for the order to finish; billable; no delete |

- A data source for each pilot resource except the number order, generated from the same descriptors.
- `catalog/coverage.yaml` entries for these operations reviewed and set to `reviewed: true`.

## Capabilities

### New Capabilities
- `codegen`: how resources are derived from the OpenAPI spec and overrides, and the behaviour every generated resource shares.
- `ai-assistants`: AI assistants, tools and MCP servers.
- `ai-avatars`: AI avatars.
- `voice`: outbound voice profiles.
- `numbers`: number orders and phone number voice settings.

### Modified Capabilities

## Impact

New `internal/generator`, `internal/engine` and `generator/overrides/` directories; generated files under `internal/provider/generated/`; docs and examples for seven resources and six data sources. Acceptance tests need `TELNYX_API_KEY`; the number order test also needs `TELNYX_ACC_BILLABLE=1` because it buys a phone number.
