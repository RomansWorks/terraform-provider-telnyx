# Proposal

## Why

The daily sync pinned `team-telnyx/openapi@434f832` (PR #13), but the pin was merged without regenerating the provider, so `main` fails the "Generated docs" check. This change accounts for every item in `api-diff.md` and regenerates.

Upstream range: `8938d09` to `434f832`.

## What Changes

- **`telnyx_ai_assistant`**: new optional, sensitive `extra_headers` (a JSON object of static HTTP headers sent to an external LLM endpoint) in `external_llm`, `delegation_settings.external_llm` and `fallback_config.external_llm`. Also in the `telnyx_ai_assistant` and `telnyx_ai_assistants` data sources.
- **`telnyx_ai_assistant`**: the `start_speaking_plan` descriptions now say how a null or empty plan behaves. No type or behaviour change.
- The other changed operations need no provider change:
  - The other AI assistant operations (list, deleted, versions, clone, import, restore, promote) changed in the same shared schemas. Their coverage entries are unchanged; they are covered by the regenerated descriptors or by later waves.
  - `GET /number_lookup/{phone_number}`, `POST /ai/chat/completions` and `POST /ai/openai/chat/completions` are `out_of_scope` (runtime operations). Their classification stands.

Nothing is removed or retyped, so there is no breaking change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The new attribute is derived from the spec by the existing codegen requirements, so no requirement changes (`skip_specs: true`).

## Impact

- `generator/overrides/ai_assistant.yaml`: `extra_headers` is sensitive.
- `internal/provider/generated/ai_assistant.go` and `ai_assistants_list.go`: regenerated.
- `docs/`: regenerated.
- A changelog entry for `extra_headers`.
