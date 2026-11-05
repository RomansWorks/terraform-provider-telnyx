## ADDED Requirements

### Requirement: Values only the create response holds
An override attribute with `create_only: true` SHALL name a top-level field that a 2xx create response holds and the read response does not. It SHALL be a computed attribute filled from the create response and kept in state by later reads. A field that no create response holds, a nested field, or a field the read response already holds SHALL be a generation error.

#### Scenario: Key shown once
- **WHEN** a `telnyx_ai_gateway_token_key` is created and the create response holds `token`
- **THEN** `token` is in state, stays there after refresh, and the next plan is empty

### Requirement: Raw bodies
A request body of `application/octet-stream` alone SHALL NOT be a generation error; the operation SHALL be marked raw, and sending it without a `Send` hook SHALL fail. A read that answers with plain text or raw bytes SHALL be read as one `value` attribute holding the body unchanged.

#### Scenario: KV value drift
- **WHEN** a KV key's value is changed outside Terraform
- **THEN** the next plan updates `value`

### Requirement: Lists in results
A list response that holds its items in `results` and their total in `count`, paged with `page` and `page_size`, SHALL be read page by page until a short page or until `count` items are read.

#### Scenario: Test agents
- **WHEN** `data "telnyx_ai_test_agents"` is read and the API reports `count = 3` with two items per page
- **THEN** two requests are sent and all three agents are returned

### Requirement: Typed lookup query parameters
A data-source-only `read` SHALL accept boolean and integer query parameters as lookup attributes, sent in their JSON form. Query parameters of other types SHALL remain generation errors.

#### Scenario: Boolean lookup
- **WHEN** `data "telnyx_ai_assistant_version"` sets `include_mcp_servers = true`
- **THEN** the request carries `include_mcp_servers=true`

### Requirement: Merged union variants
An override with `merge_variants: true` SHALL merge a union whose variants are objects with fields of the same types into one object: enums joined unless a variant takes any value, and a field Required only when every variant requires it. Without the key, such unions SHALL stay JSON attributes.

#### Scenario: One body per provider
- **WHEN** `POST /voice_clones` takes one object per provider and the override sets `merge_variants: true`
- **THEN** `name`, `language`, `gender` and `voice_provider` are attributes, and `voice_provider` accepts every provider's values

### Requirement: Request hooks
A resource's `Send` hook SHALL replace the request of its create and update, for bodies the engine cannot build. A `Headers` hook SHALL add headers to its update and delete requests.

#### Scenario: Versioned update
- **WHEN** a `telnyx_ai_gateway_token_group` at version 3 is updated
- **THEN** the PATCH carries `If-Match: "3"`

### Requirement: Generated file names
A type whose name ends in `_test` SHALL be generated into a file ending in `_test_resource.go`, so Go does not compile it as a test.

#### Scenario: Assistant tests
- **WHEN** `telnyx_ai_assistant_test` is generated
- **THEN** it is in `ai_assistant_test_resource.go` and is part of the provider
