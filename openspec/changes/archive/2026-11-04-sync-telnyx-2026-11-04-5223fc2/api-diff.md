# Telnyx API changes

Upstream `team-telnyx/openapi` from `434f832` to `5223fc2`.

| Change | Count |
|---|---|
| Operations added | 2 |
| Operations removed | 0 |
| Operations changed | 23 |
| Newly deprecated | 0 |
| Candidate resources added | 1 |
| Candidate resources removed | 0 |
| Candidate resources changed | 1 |

## Candidate resources added

Each needs a spec requirement, an override file and tests, or a coverage note explaining why not.

- `/user_profile`

## Candidate resources changed

Request or response schema changed. Regenerate and check for new, removed or retyped attributes.

- `/ai/assistants/{assistant_id}`

## Operations added

Added to coverage.yaml with reviewed: false. Confirm each bucket.

- `GET /user_profile`
- `PATCH /user_profile`

## Operations changed

Schema or parameters changed.

- `GET /ai/assistants`
- `GET /ai/assistants/deleted`
- `GET /ai/assistants/{assistant_id}`
- `GET /ai/assistants/{assistant_id}/deleted`
- `GET /ai/assistants/{assistant_id}/versions`
- `GET /ai/assistants/{assistant_id}/versions/{version_id}`
- `GET /text-to-speech/speech`
- `GET /text-to-speech/voices`
- `POST /ai/assistants`
- `POST /ai/assistants/import`
- `POST /ai/assistants/{assistant_id}`
- `POST /ai/assistants/{assistant_id}/clone`
- `POST /ai/assistants/{assistant_id}/restore`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}/promote`
- `POST /calls`
- `POST /calls/{call_control_id}/actions/answer`
- `POST /calls/{call_control_id}/actions/conversation_relay_start`
- `POST /calls/{call_control_id}/actions/gather_using_ai`
- `POST /calls/{call_control_id}/actions/gather_using_speak`
- `POST /calls/{call_control_id}/actions/speak`
- `POST /conferences/{id}/actions/speak`
- `POST /text-to-speech/speech`

