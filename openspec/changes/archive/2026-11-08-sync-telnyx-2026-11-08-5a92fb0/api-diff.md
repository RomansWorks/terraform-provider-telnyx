# Telnyx API changes

Upstream `team-telnyx/openapi` from `208f3fb` to `5a92fb0`.

| Change | Count |
|---|---|
| Operations added | 18 |
| Operations removed | 9 |
| Operations changed | 49 |
| Newly deprecated | 0 |
| Candidate resources added | 3 |
| Candidate resources removed | 2 |
| Candidate resources changed | 3 |

## Candidate resources added

Each needs a spec requirement, an override file and tests, or a coverage note explaining why not.

- `/calendar_connections/{id}`
- `/mobile_phone_numbers/{id}`
- `/mobile_voice_connections/{id}`

## Candidate resources removed

Removing a resource is a breaking change: deprecate first, remove in the next major release.

- `/v2/mobile_phone_numbers/{id}`
- `/v2/mobile_voice_connections/{id}`

## Candidate resources changed

Request or response schema changed. Regenerate and check for new, removed or retyped attributes.

- `/ai/assistants/{assistant_id}`
- `/meeting_sessions/{id}`
- `/porting_orders/{id}`

## Operations removed

Check coverage.yaml history for the Terraform type that used each one.

- `DELETE /v2/mobile_voice_connections/{id}`
- `GET /v2/mobile_phone_numbers`
- `GET /v2/mobile_phone_numbers/{id}`
- `GET /v2/mobile_voice_connections`
- `GET /v2/mobile_voice_connections/{id}`
- `PATCH /v2/mobile_phone_numbers/{id}`
- `PATCH /v2/mobile_voice_connections/{id}`
- `POST /v2/mobile_voice_connections`
- `POST /v2/payment/stored_payment_transactions`

## Operations added

Added to coverage.yaml with reviewed: false. Confirm each bucket.

- `DELETE /calendar_connections/{id}`
- `DELETE /mobile_voice_connections/{id}`
- `GET /calendar_connections`
- `GET /calendar_connections/{id}`
- `GET /calendar_connections/{id}/meetings`
- `GET /meeting_bot/calendar/oauth_callback`
- `GET /meeting_sessions/{id}/recordings`
- `GET /mobile_phone_numbers`
- `GET /mobile_phone_numbers/{id}`
- `GET /mobile_voice_connections`
- `GET /mobile_voice_connections/{id}`
- `PATCH /calendar_connections/{id}`
- `PATCH /calendar_connections/{id}/meetings/{meeting_id}`
- `PATCH /mobile_phone_numbers/{id}`
- `PATCH /mobile_voice_connections/{id}`
- `POST /calendar_connections`
- `POST /mobile_voice_connections`
- `POST /payment/stored_payment_transactions`

## Operations changed

Schema or parameters changed.

- `DELETE /agent_gateways/{instance_id}/guardrails/{guardrail_id}`
- `DELETE /meeting_sessions/{id}`
- `DELETE /meeting_sessions/{id}/recording_media`
- `GET /agent_gateways/{instance_id}/guardrails`
- `GET /ai/assistants`
- `GET /ai/assistants/deleted`
- `GET /ai/assistants/{assistant_id}`
- `GET /ai/assistants/{assistant_id}/deleted`
- `GET /ai/assistants/{assistant_id}/versions`
- `GET /ai/assistants/{assistant_id}/versions/{version_id}`
- `GET /meeting_sessions`
- `GET /meeting_sessions/{id}`
- `GET /meeting_sessions/{id}/artifacts`
- `GET /meeting_sessions/{id}/artifacts/{artifact_id}`
- `GET /meeting_sessions/{id}/events`
- `GET /meeting_sessions/{id}/transcript`
- `GET /porting/events`
- `GET /porting/events/{id}`
- `GET /porting_orders`
- `GET /porting_orders/exception_types`
- `GET /porting_orders/{id}`
- `GET /text-to-speech/speech`
- `GET /text-to-speech/voices`
- `PATCH /agent_gateways/{instance_id}/guardrails/{guardrail_id}`
- `PATCH /meeting_sessions/{id}`
- `PATCH /porting_orders/{id}`
- `POST /agent_gateways/{instance_id}/guardrails`
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
- `POST /meeting_sessions`
- `POST /meeting_sessions/{id}/actions/send_chat`
- `POST /meeting_sessions/{id}/actions/speak`
- `POST /meeting_sessions/{id}/artifacts`
- `POST /porting_orders`
- `POST /porting_orders/{id}/actions/cancel`
- `POST /porting_orders/{id}/actions/confirm`
- `POST /text-to-speech/speech`

