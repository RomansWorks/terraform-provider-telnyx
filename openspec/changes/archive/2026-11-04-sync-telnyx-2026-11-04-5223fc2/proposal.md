# Proposal

## Why

The daily sync pinned `team-telnyx/openapi@5223fc2` (PR #16), but the pin was merged without regenerating the provider, so `main` fails the "Generated docs" check again. This change accounts for every item in `api-diff.md` and regenerates.

Upstream range: `434f832` to `5223fc2`.

## What Changes

- **`telnyx_ai_assistant`**: new optional `voice_settings.language`, the two-letter language a Soniox or AssemblyAI voice speaks. Also in the `telnyx_ai_assistant` and `telnyx_ai_assistants` data sources.
- **New operations `GET /user_profile` and `PATCH /user_profile`** (beta): classified as the singleton resource `telnyx_user_profile`, built with the account wave (wave 6). No provider code yet.
- The other changed operations need no provider change:
  - The other AI assistant operations (list, deleted, versions, clone, import, restore, promote) changed in the same shared voice settings schema. Their coverage entries are unchanged.
  - `POST /calls`, the call actions (`answer`, `conversation_relay_start`, `gather_using_ai`, `gather_using_speak`, `speak`), `POST /conferences/{id}/actions/speak` and the text-to-speech operations are `out_of_scope` (runtime operations). Their classification stands.

Nothing is removed or retyped, so there is no breaking change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The new attribute is derived from the spec by the existing codegen requirements, so no requirement changes (`skip_specs: true`).

## Impact

- `catalog/coverage.yaml`: the two user profile operations are reviewed.
- `internal/provider/generated/ai_assistant.go` and `ai_assistants_list.go`: regenerated.
- `docs/`: regenerated.
- A changelog entry for `voice_settings.language`.
