# Design

## Context

A calendar connection is created by `POST /calendar_connections`, which returns a pending object with an `oauth_url`. The connection only exists once a person consents in a browser, and it is owned by that person. Terraform cannot complete that step.

## Goals / Non-Goals

**Goals:** manage the connection's policy and defaults as code; read meetings; set per-meeting overrides.

**Non-Goals:** driving the OAuth flow, or meeting session runtime data.

## Decisions

- **Adopt, like `telnyx_mobile_phone_number`.** Create is the `PATCH` with the configured fields, identified by `calendar_connection_id` (`id_from_path`). This reuses the existing engine path; no hooks.
- **`release_on_destroy`.** Disconnecting cancels future bot joins, which is drastic for an object Terraform did not create. Destroy removes the connection from state by default; `release_on_destroy = true` sends `DELETE`.
- **Auto-join override as an action.** It changes one calendar event, which Terraform does not own and which ends; a resource would drift as meetings pass.
- **Sync workflow.** `make generate` runs after `make catalog` with `continue-on-error`. If an override names an operation that upstream removed, codegen fails; the pull request still opens with the pin and the failure in its description, and the coding agent fixes the override during the sync change.

## Risks / Trade-offs

- `bot_will_join` is required on the action, so it cannot send `null` ("follow the connection policy"). Setting the connection's policy covers that case.
