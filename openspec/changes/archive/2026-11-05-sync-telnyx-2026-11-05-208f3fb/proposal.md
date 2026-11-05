# Proposal

## Why

The daily sync pinned `team-telnyx/openapi@208f3fb` (PR #18). It adds the Agent Gateway API (38 operations) and left them unreviewed in `catalog/coverage.yaml`, so the release gate (`make api-check-release`) fails. Two operations from an earlier pin (`8938d09`) were also never reviewed. This change classifies all 40.

Upstream range: `5223fc2` to `208f3fb`.

## What Changes

- **Agent Gateway (new, 38 operations)**: classified for the AI wave (wave 5). No provider code yet.
  - Resources: `telnyx_agent_gateway`, `telnyx_agent_gateway_api`, `telnyx_agent_gateway_guardrail`, `telnyx_agent_gateway_key`, `telnyx_agent_gateway_mcp_server`, `telnyx_agent_gateway_team`, `telnyx_agent_gateway_policy`, `telnyx_agent_gateway_policy_attachment`, `telnyx_agent_gateway_key_tool_permissions`, `telnyx_agent_gateway_team_tool_permissions`.
  - Data sources: `telnyx_agent_gateways`, `telnyx_agent_gateway_apis`, `telnyx_agent_gateway_guardrails`, `telnyx_agent_gateway_keys`, `telnyx_agent_gateway_mcp_servers`, `telnyx_agent_gateway_teams`, `telnyx_agent_gateway_policies`, `telnyx_agent_gateway_policy_attachments`, `telnyx_agent_gateway_models`.
  - `GET /agent_gateways/{instance_id}/spend_logs` is `out_of_scope` (reporting), like the LLM token gateway spend operations.
- **Previously unreviewed operations**:
  - `GET /telephony_credentials/{id}/token_material` returns signing secrets, so it is an ephemeral resource, `telnyx_telephony_credential_token_material`, like `telnyx_telephony_credential_token`. Deferred: see design.
  - `DELETE /ai/conversations/{conversation_id}/conversations-insights/{insight_id}` deletes stored conversation data: `out_of_scope` (runtime), like the list beside it.
- **Changed operations (19)**: the AI memory operations and `GET /ai/tests/test_framework/v1/personalities` changed only in the shared error schema (`source.parameter` was added to `TelnyxError`), and every tag gained the `agent-gateway_bearerAuth` security scheme. No attribute changes, and no generated file changes.
- **Candidate resource changed** `/ai/memory/namespaces/{namespace}/settings`: the same error schema change. `telnyx_ai_memory_namespace_settings` is not built yet (wave 5).

Nothing is removed or retyped, so there is no breaking change.

## Capabilities

### New Capabilities

None. The Agent Gateway requirements are written with wave 5.

### Modified Capabilities

None (`skip_specs: true`).

## Impact

- `catalog/coverage.yaml`: 40 operations reviewed; `make api-check-release` passes.
- No provider code, generated files or docs change, so no changelog entry.
