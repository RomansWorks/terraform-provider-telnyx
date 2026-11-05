# Design

## Context

The Agent Gateway is a new AI product: a gateway instance on a WireGuard Interface, with APIs and MCP servers it fronts, guardrails, policies and their attachments, teams, virtual keys, and per-key and per-team tool grants. The catalog found four candidate resources; the other sub-objects have create, update and delete but no single-item read.

## Goals / Non-Goals

**Goals:** classify every new operation so the release gate passes, and give wave 5 a resource plan.

**Non-Goals:** implementing any of it here. Each resource needs an override, tests and requirements, which belong to the wave 5 change.

## Decisions

- **Sub-objects without a single-item GET are resources** (`api`, `guardrail`, `key`, `mcp_server`, `team`, `policy_attachment`). They are read with `read_from` from their list operation, as `telnyx_voice_channel_zone` is. Their list operations are also list data sources.
- **`telnyx_agent_gateway` has no update operation**, so every argument forces replacement. The instance is created `pending` and becomes `ready` or `failed`; sub-objects return `409` until it is ready, so wave 5 must wait for `ready` on create.
- **`telnyx_agent_gateway_key`**: the key value is only in the create response. Wave 5 stores it as a sensitive computed attribute (the only way users can get it), and should consider an ephemeral alternative.
- **Tool permissions are singletons**: `PUT` replaces the grant. Destroy sets `servers` to `[]`, which the API documents as granting none.
- **Spend logs are out of scope** as reporting, consistent with `GET /llm_token_gateway/spend/*`.
- **`telnyx_telephony_credential_token_material`** is classified `ephemeral` so the secrets never reach state. It is a follow-up to wave 1 (telephony credentials), which has already merged; it is not assigned to a wave and is noted in the pull request.

## Risks / Trade-offs

- The Agent Gateway API is new and may change before wave 5 builds it; the next sync will show any change.
