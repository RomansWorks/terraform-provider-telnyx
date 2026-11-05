## ADDED Requirements

### Requirement: Assistant tags
The provider SHALL manage `telnyx_ai_assistant_tag`, one tag on an assistant, added with `POST /ai/assistants/{assistant_id}/tags` and removed with `DELETE /ai/assistants/{assistant_id}/tags/{tag}`. It SHALL be read from the assistant's `tags`; a tag missing there SHALL be treated as deleted outside Terraform. Every argument SHALL force replacement. Import SHALL take `<assistant_id>/<tag>`.

#### Scenario: Lifecycle
- **WHEN** a tag is created, refreshed, changed, destroyed and imported
- **THEN** it is added, the next plan is empty, a new tag replaces it, it is removed, and import yields an empty plan

#### Scenario: Removed outside Terraform
- **WHEN** the tag disappears from the assistant's `tags`
- **THEN** the next apply adds it again

### Requirement: Assistant canary deployments, tests and versions
The provider SHALL manage `telnyx_ai_assistant_canary_deploy` (at most one per assistant, identified and imported by `assistant_id`) and `telnyx_ai_assistant_test` (identified by `test_id`).

#### Scenario: Canary lifecycle
- **WHEN** a canary deployment is created, its rules changed, destroyed and imported by assistant id
- **THEN** the matching POST, PUT, DELETE and GET are sent, and every plan after apply is empty

#### Scenario: Promote a version
- **WHEN** `action "telnyx_ai_assistant_version_promote"` is invoked
- **THEN** `POST /ai/assistants/{assistant_id}/versions/{version_id}/promote` is sent once

### Requirement: Assistant data sources and actions
The provider SHALL offer data sources for assistant versions, deleted assistants, tag names and test suites, and actions to clone, import and restore an assistant and to update, promote or delete a version. Clone and import SHALL say that Terraform does not manage what they create. `provider` of the import action SHALL be `source_provider`.

#### Scenario: Import an assistant
- **WHEN** `action "telnyx_ai_assistant_import"` is invoked with `source_provider = "vapi"`
- **THEN** the request body carries `provider: "vapi"`

### Requirement: Integration secrets
The provider SHALL manage `telnyx_integration_secret`, read from `GET /integration_secrets`. It has no update: every argument SHALL force replacement. `token` and `password` SHALL be sensitive and, since the API never returns them, kept from configuration.

#### Scenario: Rotate a token
- **WHEN** `token` changes
- **THEN** the plan replaces the secret

### Requirement: Tool attachments are set on the assistant
Attaching shared tools to an assistant SHALL be done with `tool_ids` on `telnyx_ai_assistant`; `PUT` and `DELETE /ai/assistants/{assistant_id}/tools/{tool_id}` SHALL be out of scope, because the assistant's read does not return attached tool ids.

#### Scenario: Attach a tool
- **WHEN** `tool_ids` on an assistant lists a shared tool's id
- **THEN** the assistant update sends it
