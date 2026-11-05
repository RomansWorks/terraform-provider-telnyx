# Telnyx API changes

Upstream `team-telnyx/openapi` from `5223fc2` to `208f3fb`.

| Change | Count |
|---|---|
| Operations added | 38 |
| Operations removed | 0 |
| Operations changed | 19 |
| Newly deprecated | 0 |
| Candidate resources added | 4 |
| Candidate resources removed | 0 |
| Candidate resources changed | 1 |

## Candidate resources added

Each needs a spec requirement, an override file and tests, or a coverage note explaining why not.

- `/agent_gateways/{instance_id}`
- `/agent_gateways/{instance_id}/keys/{key_id}/tool_permissions`
- `/agent_gateways/{instance_id}/policies/{name}`
- `/agent_gateways/{instance_id}/teams/{team_id}/tool_permissions`

## Candidate resources changed

Request or response schema changed. Regenerate and check for new, removed or retyped attributes.

- `/ai/memory/namespaces/{namespace}/settings`

## Operations added

Added to coverage.yaml with reviewed: false. Confirm each bucket.

- `DELETE /agent_gateways/{instance_id}`
- `DELETE /agent_gateways/{instance_id}/apis/{api_id}`
- `DELETE /agent_gateways/{instance_id}/guardrails/{guardrail_id}`
- `DELETE /agent_gateways/{instance_id}/keys/{key_id}`
- `DELETE /agent_gateways/{instance_id}/mcp_servers/{server_id}`
- `DELETE /agent_gateways/{instance_id}/policies/{name}`
- `DELETE /agent_gateways/{instance_id}/policy_attachments/{attachment_id}`
- `DELETE /agent_gateways/{instance_id}/teams/{team_id}`
- `GET /agent_gateways`
- `GET /agent_gateways/{instance_id}`
- `GET /agent_gateways/{instance_id}/apis`
- `GET /agent_gateways/{instance_id}/guardrails`
- `GET /agent_gateways/{instance_id}/keys`
- `GET /agent_gateways/{instance_id}/keys/{key_id}/tool_permissions`
- `GET /agent_gateways/{instance_id}/mcp_servers`
- `GET /agent_gateways/{instance_id}/models`
- `GET /agent_gateways/{instance_id}/policies`
- `GET /agent_gateways/{instance_id}/policies/{name}`
- `GET /agent_gateways/{instance_id}/policy_attachments`
- `GET /agent_gateways/{instance_id}/spend_logs`
- `GET /agent_gateways/{instance_id}/teams`
- `GET /agent_gateways/{instance_id}/teams/{team_id}/tool_permissions`
- `PATCH /agent_gateways/{instance_id}/apis/{api_id}`
- `PATCH /agent_gateways/{instance_id}/guardrails/{guardrail_id}`
- `PATCH /agent_gateways/{instance_id}/keys/{key_id}`
- `PATCH /agent_gateways/{instance_id}/mcp_servers/{server_id}`
- `PATCH /agent_gateways/{instance_id}/policies/{name}`
- `PATCH /agent_gateways/{instance_id}/teams/{team_id}`
- `POST /agent_gateways`
- `POST /agent_gateways/{instance_id}/apis`
- `POST /agent_gateways/{instance_id}/guardrails`
- `POST /agent_gateways/{instance_id}/keys`
- `POST /agent_gateways/{instance_id}/mcp_servers`
- `POST /agent_gateways/{instance_id}/policies`
- `POST /agent_gateways/{instance_id}/policy_attachments`
- `POST /agent_gateways/{instance_id}/teams`
- `PUT /agent_gateways/{instance_id}/keys/{key_id}/tool_permissions`
- `PUT /agent_gateways/{instance_id}/teams/{team_id}/tool_permissions`

## Operations changed

Schema or parameters changed.

- `DELETE /ai/memory/namespaces/{namespace}`
- `DELETE /ai/memory/namespaces/{namespace}/profiles/{profile_id}`
- `DELETE /ai/memory/namespaces/{namespace}/profiles/{profile_id}/sources/{source_id}`
- `GET /ai/memory/namespaces`
- `GET /ai/memory/namespaces/{namespace}/operations/{operation_id}`
- `GET /ai/memory/namespaces/{namespace}/profiles`
- `GET /ai/memory/namespaces/{namespace}/profiles/{profile_id}/memories`
- `GET /ai/memory/namespaces/{namespace}/profiles/{profile_id}/memories/{memory_id}`
- `GET /ai/memory/namespaces/{namespace}/profiles/{profile_id}/sources`
- `GET /ai/memory/namespaces/{namespace}/profiles/{profile_id}/sources/{source_id}`
- `GET /ai/memory/namespaces/{namespace}/profiles/{profile_id}/summary`
- `GET /ai/memory/namespaces/{namespace}/settings`
- `GET /ai/tests/test_framework/v1/personalities`
- `PATCH /ai/memory/namespaces/{namespace}/settings`
- `POST /ai/memory/namespaces`
- `POST /ai/memory/namespaces/{namespace}/profiles/{profile_id}/ingest`
- `POST /ai/memory/namespaces/{namespace}/profiles/{profile_id}/recall`
- `POST /ai/memory/namespaces/{namespace}/profiles/{profile_id}/remember`
- `POST /ai/memory/namespaces/{namespace}/profiles/{profile_id}/summary/preview`

