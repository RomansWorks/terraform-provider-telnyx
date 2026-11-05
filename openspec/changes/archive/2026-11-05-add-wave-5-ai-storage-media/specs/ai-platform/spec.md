## ADDED Requirements

### Requirement: Knowledge collections
The provider SHALL manage `telnyx_ai_collection`, identified by `uuid`, and `telnyx_ai_collection_source`, read from the collection's source list and imported as `<collection_id>/<id>`. A collection's `retrieval` settings SHALL be sent with `PATCH /ai/collections/{uuid}/settings`, also straight after create, and read back from `settings.retrieval`. `sources` on the collection SHALL only report sources.

#### Scenario: Change retrieval
- **WHEN** `retrieval.top_k` changes from 5 to 8
- **THEN** the provider sends `PATCH /ai/collections/{uuid}/settings` and the next plan is empty

#### Scenario: Collection lifecycle
- **WHEN** a collection is created, renamed, destroyed and imported
- **THEN** each step uses the matching operation and import yields an empty plan

### Requirement: Collection lookups
`data "telnyx_ai_collection_by_slug"` SHALL look up a collection by its slug, and `data "telnyx_ai_knowledge_collection_documents"` SHALL list a collection's documents.

#### Scenario: Look up by slug
- **WHEN** the data source sets `slug`
- **THEN** the provider reads the collection with that slug

### Requirement: Memory namespaces
The provider SHALL manage `telnyx_ai_memory_namespace`, identified by `name` (the paths use it) and read from the namespace list, and `telnyx_ai_memory_namespace_settings`, identified by the namespace. Destroying the settings SHALL remove them from state only.

#### Scenario: Namespace gone
- **WHEN** a namespace is no longer listed
- **THEN** it is removed from state

### Requirement: Missions
The provider SHALL manage `telnyx_ai_mission`, identified by `mission_id`, and offer the `telnyx_ai_mission_clone` action. Mission knowledge bases, MCP servers and tools SHALL be out of scope until the spec gives their schemas.

#### Scenario: Mission lifecycle
- **WHEN** a mission is created, its instructions changed with PUT, destroyed and imported
- **THEN** each step uses the matching operation and every plan after apply is empty

### Requirement: Conversation insights
The provider SHALL manage `telnyx_ai_conversation_insight`, `telnyx_ai_conversation_insight_group` and `telnyx_ai_conversation_insight_group_assignment`. An assignment SHALL be read from its group's `insights`, identified by the insight id, imported as `<group_id>/<insight_id>`, and SHALL not repeat the insight's own attributes.

#### Scenario: Assign
- **WHEN** an assignment is created
- **THEN** `POST /ai/conversations/insight-groups/{group_id}/insights/{insight_id}/assign` is sent and the next plan is empty

### Requirement: AI test framework
The provider SHALL manage `telnyx_ai_test_agent`, `telnyx_ai_test_metric`, `telnyx_ai_test_scenario` and `telnyx_ai_test_schedule`, and offer their list data sources and `telnyx_ai_test_personalities`. Lists that answer with `results` SHALL be paged. The metric bulk create SHALL be out of scope.

#### Scenario: Scenario lifecycle
- **WHEN** a scenario is created, its instructions changed with PATCH, destroyed and imported by its integer id
- **THEN** each step uses the matching operation and every plan after apply is empty

### Requirement: AI gateway tokens
The provider SHALL manage `telnyx_ai_gateway_token_group` and `telnyx_ai_gateway_token_key`. Update and delete SHALL send `If-Match` with the quoted `version` from state, and a plan that changes anything else SHALL leave `version` and `updated_at` unknown. A key's `token` SHALL be sensitive and kept from the create response. `allowed_models` and `provider_key_ids` SHALL be sets.

#### Scenario: Changed outside Terraform
- **WHEN** a token group was changed in the portal since the last refresh and Terraform updates it
- **THEN** the API rejects the stale `If-Match` and apply fails instead of overwriting the change

#### Scenario: Token kept
- **WHEN** a token key is created and later renamed
- **THEN** `token` holds the created token throughout and the PATCH carries `If-Match: "1"`
