# Proposal

## Why

The daily sync pinned `team-telnyx/openapi@5a92fb0` (PR #26). It adds meeting bot calendar connections, fixes the mobile phone number and mobile voice connection paths that repeated `/v2`, and changes 49 operations. The new operations are unreviewed, so the release gate fails.

Upstream range: `208f3fb` to `5a92fb0`.

Sync pull requests have also broken the `Generated docs` check on `main` twice, because the daily workflow pins the spec without regenerating descriptors and docs.

## What Changes

- **Meeting bot calendar connections (new, 8 operations)**:
  - `telnyx_calendar_connection` manages a connection's auto-join policy and future-session defaults. A connection only exists after the user finishes Google or Microsoft OAuth in a browser, so create adopts an existing connection (`PATCH /calendar_connections/{id}`), and destroy removes it from state unless `release_on_destroy = true`, which disconnects it (`DELETE`). Data sources `telnyx_calendar_connection` and `telnyx_calendar_connections`.
  - `data "telnyx_calendar_meetings"` lists a connection's tracked meetings; the action `telnyx_calendar_meeting_auto_join` sets one meeting's override.
  - `POST /calendar_connections` (starts OAuth, returns `oauth_url`) and `GET /meeting_bot/calendar/oauth_callback` (browser redirect) are `out_of_scope`: OAuth flows are runtime operations.
- **Mobile paths moved (9 removed, 8 added)**: the `/v2/mobile_phone_numbers` and `/v2/mobile_voice_connections` operations now live at `/mobile_...`. `telnyx_mobile_phone_number`, `telnyx_mobile_voice_connection` and their list data sources move to the new paths. The request URLs are the same (the client already sent `/v2/...` once), so users see no change.
- **`POST /payment/stored_payment_transactions`** replaces the `/v2/` path: still `out_of_scope` (runtime: payments).
- **`GET /meeting_sessions/{id}/recordings`**: `out_of_scope` (runtime: meeting data), like the other meeting session operations.
- **Changed operations (49)** and candidate resources changed:
  - Porting orders and exception types: exceptions gain `remediation` (computed, additive) in `telnyx_porting_order`, `telnyx_porting_orders` and `telnyx_porting_order_exception_types`.
  - AI assistants: description changes only (`voice_settings.language`, `in_transit_data_locality`).
  - Meeting sessions, calls, conferences and text-to-speech are `out_of_scope`; the agent gateway guardrail changes leave the generated descriptors unchanged.
- **Daily sync workflow** runs `make generate` too, so its pull request carries regenerated descriptors and docs.

Nothing is removed or retyped, so there is no breaking change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-platform`: meeting bot calendar connections.
- `codegen`: the versioned-path scenario no longer names mobile phone numbers, whose paths no longer repeat `/v2`.
- `api-sync`: the scheduled sync regenerates descriptors and docs.

## Impact

- `generator/overrides/`: three new files, two path updates; `internal/provider/generated/` and `docs/` regenerated.
- `catalog/coverage.yaml`: 18 added operations reviewed.
- `.github/workflows/sync.yml`: runs `make generate` and commits its output.
- Changelog: two features, one enhancement.
