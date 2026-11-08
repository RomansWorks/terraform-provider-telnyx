# ai-assistants Specification

## Purpose
Manages Telnyx AI assistants and the tools and MCP servers they use.

## Requirements

### Requirement: AI assistant
The provider SHALL manage `telnyx_ai_assistant` with `POST /ai/assistants`, `GET /ai/assistants/{assistant_id}`, `POST /ai/assistants/{assistant_id}` (update) and `DELETE /ai/assistants/{assistant_id}`. `tags` SHALL be a set. `conversation_flow` and `dynamic_variables` SHALL be JSON strings.

#### Scenario: Create and read back
- **WHEN** a user applies an assistant with `name`, `model` and `instructions`
- **THEN** the assistant is created and the next plan is empty

#### Scenario: Change instructions
- **WHEN** `instructions` changes
- **THEN** the assistant is updated in place with POST

#### Scenario: Destroy and import
- **WHEN** the assistant is destroyed, or imported by id
- **THEN** it is deleted, or read into state with an empty follow-up plan

### Requirement: AI tool
The provider SHALL manage `telnyx_ai_tool` with `POST /ai/tools`, `GET`, `PATCH` and `DELETE /ai/tools/{tool_id}`. Exactly one tool variant attribute (`webhook`, `handoff`, `retrieval`, `function`, `invite`, `pay`, `client_side_tool`, `update_dynamic_variables`, `vision`) SHALL be set, and it SHALL match `type`; variants without settings are set to `{}`. Read SHALL map `tool_definition` back onto the variant so that drift is detected. Changing `type` SHALL force replacement.

#### Scenario: Webhook URL changed in the portal
- **WHEN** the webhook URL is changed outside Terraform
- **THEN** the next plan shows the URL change

#### Scenario: Type mismatch
- **WHEN** `type` is `webhook` but the `handoff` variant is set
- **THEN** validation fails before any API call

#### Scenario: Lifecycle
- **WHEN** a tool is created, updated, destroyed and imported
- **THEN** each step uses the matching operation, and import reads the variant's settings from `tool_definition`

#### Scenario: Type change
- **WHEN** `type` changes
- **THEN** the plan replaces the tool

### Requirement: AI MCP server
The provider SHALL manage `telnyx_ai_mcp_server` with `POST /ai/mcp_servers`, `GET`, `PUT` and `DELETE /ai/mcp_servers/{mcp_server_id}`. Updates SHALL send the full object. `allowed_tools` SHALL be a set.

#### Scenario: Update sends everything
- **WHEN** only `name` changes
- **THEN** the PUT request still carries `type`, `url` and the other configured attributes

#### Scenario: Lifecycle
- **WHEN** an MCP server is created, destroyed and imported
- **THEN** each step uses the matching operation and import yields an empty plan

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
