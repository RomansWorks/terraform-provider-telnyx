# Proposal

## Why

The daily sync pinned `team-telnyx/openapi@bf2eb8e` (branch `sync/telnyx-api`). Upstream range: `5a92fb0` to `bf2eb8e`. It removes the AI fine-tuning job operations and changes 27 operations.

A coverage review on 2026-11-11 also found three operations classified `out_of_scope` that users configure rather than run: TeXML secrets, video room client tokens and storage presigned URLs. This change implements them.

## What Changes

- **Removed operations (4)**: `GET /ai/fine_tuning/jobs`, `POST /ai/fine_tuning/jobs`, `GET /ai/fine_tuning/jobs/{job_id}` and `POST /ai/fine_tuning/jobs/{job_id}/cancel`. They were `out_of_scope` (runtime: model training jobs) and no Terraform type used them; `make catalog` drops their coverage entries.
- **Changed operations (27)** and candidate resources changed, after `make generate`:
  - `telnyx_network_route`: `type` also accepts `wireguard_peer` (additive enum value).
  - `telnyx_ai_assistant`, its data sources and `telnyx_ai_assistant_version_update`: transcription `model` also accepts `soniox/stt-rt-v6`; Soniox setting descriptions now name it.
  - `telnyx_enterprise_reputation_remediation`, `telnyx_texml_applications`: description changes only.
  - Calls, TeXML calls, speech-to-text and the remaining AI model lists are `out_of_scope` and unchanged in coverage.
- **Coverage gaps found in review**:
  - `telnyx_texml_secret` (new resource, `POST /texml/secrets`): a named value TeXML templates reference. The API creates or replaces a secret and can neither read nor delete it. `value_wo` is write-only; `name` and `value_wo_version` force replacement; refresh keeps state; import is refused; destroy removes it from state and warns.
  - `telnyx_room_client_token` (new ephemeral, `POST /rooms/{room_id}/actions/generate_join_client_token`), with optional `token_ttl_secs` and `refresh_token_ttl_secs`.
  - `telnyx_storage_object_presigned_url` (new ephemeral, `POST /storage/buckets/{bucketName}/{objectName}/presigned_url`), with optional `ttl`.
- **Codegen**: `read: none` for resources the API cannot read back; ephemeral resources take their operation's request body fields as optional arguments.

Nothing is removed or retyped, so there is no breaking change.

## Deferred

- AI mission knowledge bases, MCP servers and tools (15 operations): the spec still gives no request or response schema. They stay `out_of_scope` until it does.
- `POST /rooms/{room_id}/actions/refresh_client_token`: takes the refresh token in its body and needs no API key; a client refreshes its own token at runtime. Stays `out_of_scope`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `codegen`: unreadable resources; arguments for ephemeral resources.
- `voice`: TeXML secrets.
- `media`: room client tokens.
- `storage`: presigned object URLs.

## Impact

- `generator/overrides/`: three new files; `internal/codegen` (override validation, derive, emit) and `internal/engine` (resource read and import, ephemeral open) with tests; conformance test refuses import for unreadable resources.
- `catalog/coverage.yaml`: fine-tuning entries removed; three entries move from `out_of_scope` to `resource` and `ephemeral`.
- `internal/provider/generated/`, `docs/` regenerated; three examples; changelog: two features, one enhancement.
