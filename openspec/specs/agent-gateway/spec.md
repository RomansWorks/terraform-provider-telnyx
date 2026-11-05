# agent-gateway Specification

## Purpose
How the provider manages Telnyx Agent Gateways (LLM and tool gateways on a WireGuard Interface) and their APIs, MCP servers, guardrails, policies, keys, teams and tool permissions.

## Requirements

### Requirement: Agent Gateways
The provider SHALL manage `telnyx_agent_gateway`. It has no update, so every argument SHALL force replacement. Create SHALL wait, within the `create` timeout (10 minutes by default), until the gateway's `status` is `ready`, and SHALL fail when it becomes `failed` or `deleting`; the gateway stays in state so that the next apply replaces it.

#### Scenario: Wait for ready
- **WHEN** `POST /agent_gateways` answers `status = "pending"` and a later read answers `ready`
- **THEN** create returns after that read with `gateway_url` set

#### Scenario: Gateway failed to start
- **WHEN** the gateway's `status` becomes `failed`
- **THEN** the apply fails naming the gateway, and it stays in state

### Requirement: Agent Gateway sub-objects
The provider SHALL manage `telnyx_agent_gateway_api`, `_guardrail`, `_key`, `_mcp_server`, `_team` and `_policy_attachment`, each read from its gateway's list and imported as `<gateway_id>/<id>`, and `telnyx_agent_gateway_policy`, read by name and imported as `<gateway_id>/<name>`. A policy attachment has no update, so its arguments SHALL force replacement. The list operations SHALL also be list data sources, with `data "telnyx_agent_gateway_models"`.

#### Scenario: Removed outside Terraform
- **WHEN** a key is no longer in `GET /agent_gateways/{instance_id}/keys`
- **THEN** it leaves state and the next plan creates it again

### Requirement: Agent Gateway secrets
A key's `key` SHALL be sensitive and kept from the create response, the only response that holds it; it is null after import. The outbound `auth` of APIs and MCP servers and a guardrail's `settings` SHALL be write-only (`auth_wo`, `settings_wo`, each with a `_wo_version`), because responses leave their credentials out. `static_headers` SHALL be sensitive.

#### Scenario: Rotate an MCP server credential
- **WHEN** `auth_wo` changes and `auth_wo_version` goes from 1 to 2
- **THEN** the provider sends the new `auth` with `PATCH .../mcp_servers/{server_id}`

### Requirement: Agent Gateway tool permissions
The provider SHALL manage `telnyx_agent_gateway_key_tool_permissions` and `telnyx_agent_gateway_team_tool_permissions`, one per key or team, imported as `<gateway_id>/<key_id>` or `<gateway_id>/<team_id>`. Create and update SHALL replace the grant with `PUT`, `servers` SHALL be a set, and destroy SHALL write `{"servers": []}`, which grants none.

#### Scenario: Destroy grants none
- **WHEN** a key's tool permissions are destroyed
- **THEN** the provider sends `PUT .../keys/{key_id}/tool_permissions` with `{"servers":[]}`
