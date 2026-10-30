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
