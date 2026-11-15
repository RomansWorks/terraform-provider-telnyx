# Telnyx API changes

Upstream `team-telnyx/openapi` from `5a92fb0` to `bf2eb8e`.

| Change | Count |
|---|---|
| Operations added | 0 |
| Operations removed | 4 |
| Operations changed | 27 |
| Newly deprecated | 0 |
| Candidate resources added | 0 |
| Candidate resources removed | 0 |
| Candidate resources changed | 4 |

## Candidate resources changed

Request or response schema changed. Regenerate and check for new, removed or retyped attributes.

- `/ai/assistants/{assistant_id}`
- `/network_routes/{id}`
- `/texml/Accounts/{account_sid}/Calls/{call_sid}`
- `/texml_applications/{id}`

## Operations removed

Check coverage.yaml history for the Terraform type that used each one.

- `GET /ai/fine_tuning/jobs`
- `GET /ai/fine_tuning/jobs/{job_id}`
- `POST /ai/fine_tuning/jobs`
- `POST /ai/fine_tuning/jobs/{job_id}/cancel`

## Operations changed

Schema or parameters changed.

- `GET /ai/assistants`
- `GET /ai/assistants/deleted`
- `GET /ai/assistants/{assistant_id}`
- `GET /ai/assistants/{assistant_id}/deleted`
- `GET /ai/assistants/{assistant_id}/versions`
- `GET /ai/assistants/{assistant_id}/versions/{version_id}`
- `GET /ai/models`
- `GET /ai/openai/models`
- `GET /enterprises/{enterprise_id}/reputation/remediation/{remediation_id}`
- `GET /speech-to-text/transcription`
- `GET /texml_applications`
- `POST /ai/assistants`
- `POST /ai/assistants/import`
- `POST /ai/assistants/{assistant_id}`
- `POST /ai/assistants/{assistant_id}/clone`
- `POST /ai/assistants/{assistant_id}/restore`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}`
- `POST /ai/assistants/{assistant_id}/versions/{version_id}/promote`
- `POST /calls`
- `POST /calls/{call_control_id}/actions/ai_assistant_add_messages`
- `POST /calls/{call_control_id}/actions/ai_assistant_start`
- `POST /calls/{call_control_id}/actions/answer`
- `POST /calls/{call_control_id}/actions/gather_using_ai`
- `POST /calls/{call_control_id}/actions/transcription_start`
- `POST /enterprises/{enterprise_id}/reputation/remediation`
- `POST /network_routes`
- `POST /texml/Accounts/{account_sid}/Calls`

