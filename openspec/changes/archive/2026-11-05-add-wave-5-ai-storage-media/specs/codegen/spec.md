## ADDED Requirements

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

### Requirement: Parent names in list data sources
A list data source's path parameters SHALL take the names the resource's `path_params` give them, except a rename to `id`, which names the item itself.

#### Scenario: Keys of a namespace
- **WHEN** `telnyx_storage_kv_key` maps path parameter `id` to `namespace_id`
- **THEN** `data "telnyx_storage_kv_keys"` takes `namespace_id`

### Requirement: Snapshot paging
A `page[number]` list SHALL send the first page's `meta.snapshot` as `page[snapshot]` on later pages, and SHALL stop when `meta.has_more` is false.

#### Scenario: AI gateway keys over two pages
- **WHEN** page 1 of `GET /llm_token_gateway/token_keys` answers `meta.snapshot = "s1"` and `has_more = true`
- **THEN** page 2 is requested with `page[snapshot]=s1`

### Requirement: Text reads are UTF-8
A text read SHALL fail with an error naming the operation when the response is not valid UTF-8, since Terraform strings cannot hold it.

#### Scenario: Binary value
- **WHEN** a KV key written outside Terraform holds bytes that are not UTF-8
- **THEN** refresh fails naming `GET /storage/kvs/{id}/keys/{key}`

## MODIFIED Requirements

### Requirement: Values only create returns
An attribute override MAY set `create_only: true` for a top-level response field that no request takes. The attribute SHALL be computed and sensitive. When the read response schema lacks the field, the generator SHALL take it from the create response schemas, looking at every 2xx response from the lowest code up (a create can answer 200 for a replay without the value and 201 with it), and SHALL fail when none lists it. The engine SHALL fill the value from the create response and SHALL keep the value in state on every read.

#### Scenario: Token only create returns
- **WHEN** `telnyx_tunnel` is created and `POST /tunnels` answers with `tunnel_token`
- **THEN** `tunnel_token` is in state, shown as sensitive
- **WHEN** the tunnel refreshes and `GET /tunnels/{id}` has no `tunnel_token`
- **THEN** the value in state is unchanged and the plan is empty

#### Scenario: Read answers null
- **WHEN** a WireGuard peer refreshes and `GET /wireguard_peers/{id}` answers `"private_key": null`
- **THEN** the private key from create stays in state

#### Scenario: Value only in the 201 response
- **WHEN** `POST /llm_token_gateway/token_keys` answers 200 without `token` and 201 with it
- **THEN** `telnyx_ai_gateway_token_key` has a sensitive computed `token` taken from the 201 schema

### Requirement: Objects without delete
When the API offers no delete for an object and the resource has no `Delete` hook, destroy SHALL remove it from state and warn that the object still exists in Telnyx. A `Delete` hook SHALL destroy such an object, for example by writing an empty grant.

#### Scenario: Destroy without delete
- **WHEN** a resource with no delete operation is destroyed
- **THEN** no API call is made, the object leaves state, and a warning names it

#### Scenario: Destroy by a Delete hook
- **WHEN** `telnyx_agent_gateway_key_tool_permissions` is destroyed
- **THEN** the provider sends `PUT .../tool_permissions` with `{"servers":[]}` and no warning
