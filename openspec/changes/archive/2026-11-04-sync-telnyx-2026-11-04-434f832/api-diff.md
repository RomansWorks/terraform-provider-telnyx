# Telnyx API changes

Upstream `team-telnyx/openapi` from `8938d09` to `434f832`.

| Change | Count |
|---|---|
| Operations added | 0 |
| Operations removed | 0 |
| Operations changed | 16 |
| Newly deprecated | 0 |
| Candidate resources added | 0 |
| Candidate resources removed | 0 |
| Candidate resources changed | 1 |

## Candidate resources changed

Request or response schema changed. Regenerate and check for new, removed or retyped attributes.

- `/ai/assistants/{assistant_id}`

## Operations changed

Schema or parameters changed.

- `GET /ai/assistants`
- `GET /ai/assistants/deleted`
- `GET /ai/assistants/{assistant_id}`
- `GET /ai/assistants/{assistant_id}/deleted`
- `GET /ai/assistants/{assistant_id}/versions`
- `GET /ai/assistants/{assistant_id}/versions/{version_id}`
- `GET /number_lookup/{phone_number}`
- `POST /ai/assistants`
- `POST /ai/assistants/import`
- `POST /ai/assistants/{assistant_id}`
- `POST /ai/assistants/{assistant_id}/clone`
- `POST /ai/assistants/{assistant_id}/restore`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}/promote`
- `POST /ai/chat/completions`
- `POST /ai/openai/chat/completions`

