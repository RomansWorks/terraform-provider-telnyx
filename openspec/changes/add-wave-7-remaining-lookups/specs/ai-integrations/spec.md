## ADDED Requirements

### Requirement: AI integrations
The provider SHALL offer read-only data sources for the integrations AI assistants can use (`telnyx_ai_integration`, `telnyx_ai_integrations`) and the account's connections to them (`telnyx_ai_integration_connection`, `telnyx_ai_integration_connections`), and the action `telnyx_ai_integration_connection_delete`, which deletes a connection once when invoked.

#### Scenario: Read an integration
- **WHEN** `data "telnyx_ai_integration" "i" { integration_id = "hubspot" }` is read
- **THEN** the provider sends `GET /ai/integrations/hubspot`

#### Scenario: Delete a connection
- **WHEN** `telnyx_ai_integration_connection_delete` is invoked with `user_connection_id = "uc1"`
- **THEN** the provider sends `DELETE /ai/integrations/connections/uc1` once
