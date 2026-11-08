# codegen Specification

## Purpose
Defines how Terraform resources and data sources are derived from the pinned Telnyx OpenAPI spec and per-resource overrides, and the behaviour all generated resources share.

## Requirements

### Requirement: Resources are generated from the spec and overrides
Each resource SHALL be defined by an override file in `generator/overrides/` naming its Terraform type and its operations as `METHOD /path`. Its attributes SHALL come from the operations' request and response schemas in the pinned spec, adjusted only by the override.

#### Scenario: Regeneration is deterministic
- **WHEN** `make generate` runs twice with no change to the spec or overrides
- **THEN** the second run changes no file

#### Scenario: Upstream adds a field
- **WHEN** a sync adds an optional field to a resource's request and response schemas
- **THEN** `make generate` adds an optional attribute for it with no override change

### Requirement: Attribute derivation
Attributes SHALL be Required when required by the create request, Optional and Computed when optional in the request and returned by read, Computed when only returned, and kept from configuration when sent but never returned. Attributes missing from every update request SHALL force replacement.

#### Scenario: Server default
- **WHEN** a user omits an optional attribute and the API returns a default value
- **THEN** the plan after apply is empty

#### Scenario: Update-only field
- **WHEN** a user sets an attribute that only the update operation accepts
- **THEN** the provider sends it in an update straight after create

#### Scenario: Immutable field
- **WHEN** a user changes an attribute that no update operation accepts
- **THEN** the plan replaces the resource

### Requirement: Polymorphic and free-form values
Unions whose object variants each fix `type` to one value SHALL become one nested object with the union of the variants' fields, where `type` selects the variant. Free-form objects and other unions SHALL become strings holding JSON, compared after normalisation. A JSON attribute SHALL keep its configured value when the API returns the same document with extra keys.

#### Scenario: Tagged union
- **WHEN** an assistant configures `tools = [{ type = "hangup", hangup = {} }]`
- **THEN** the request carries `{"type": "hangup", "hangup": {}}`

#### Scenario: JSON formatting
- **WHEN** a JSON attribute differs from the API value only in key order or whitespace
- **THEN** the plan shows no change

#### Scenario: Server defaults in JSON
- **WHEN** the API returns a JSON attribute with keys the user did not set, and the keys the user set are unchanged
- **THEN** the plan shows no change

### Requirement: Response envelopes
When a response schema has a `data` property that is an object, the provider SHALL read the object from `data`, whatever other properties the response has. When a create operation's `data` is a list, the provider SHALL take its only element as the created object, and SHALL fail naming the ids when there is more than one. A single-object read whose `data` is a list SHALL be a generation error.

#### Scenario: Wrapped response
- **WHEN** the API returns `{"data": {...}}`
- **THEN** the resource's attributes come from the inner object

#### Scenario: Batch create
- **WHEN** `POST /porting_orders` answers `{"data": [{"id": "po1", ...}]}`
- **THEN** the porting order `po1` is in state
- **WHEN** it answers with two orders
- **THEN** apply fails and the error names both ids

#### Scenario: Wrapped response with metadata
- **WHEN** the API returns `{"data": {...}, "meta": {...}}`
- **THEN** the resource's attributes come from `data`, and `meta` is ignored

### Requirement: Shared lifecycle behaviour
Every generated resource SHALL remove itself from state when read returns 404, support import by id (or `parent_id/id` for nested paths), and send only changed attributes on PATCH or POST updates and the full object on PUT updates.

#### Scenario: Deleted outside Terraform
- **WHEN** an object was deleted in the Telnyx portal
- **THEN** the next plan proposes to create it again

#### Scenario: Import
- **WHEN** a user runs `terraform import telnyx_ai_assistant.x <id>`
- **THEN** the resource is read into state and the next plan is empty for a matching configuration

### Requirement: Objects without delete
When the API offers no delete for an object and the resource has no `Delete` hook, destroy SHALL remove it from state and warn that the object still exists in Telnyx. A `Delete` hook SHALL destroy such an object, for example by writing an empty grant.

#### Scenario: Destroy without delete
- **WHEN** a resource with no delete operation is destroyed
- **THEN** no API call is made, the object leaves state, and a warning names it

#### Scenario: Destroy by a Delete hook
- **WHEN** `telnyx_agent_gateway_key_tool_permissions` is destroyed
- **THEN** the provider sends `PUT .../tool_permissions` with `{"servers":[]}` and no warning

### Requirement: Data sources
A resource's read operation SHALL also be exposed as a data source of the same name that looks the object up by id, with every attribute Computed.

#### Scenario: Look up by id
- **WHEN** a configuration declares `data "telnyx_ai_assistant" "x" { id = "..." }`
- **THEN** its attributes are populated from `GET /ai/assistants/{assistant_id}`

### Requirement: List data sources
A resource override's `list` operation, or a data-source-only override, SHALL produce a data source with:
- optional filter attributes derived from the operation's query parameters, excluding paging
- an optional `max_items`
- one computed list attribute of the items

The engine SHALL fetch pages until the last page or `max_items`, and SHALL send only the filters that are set.

#### Scenario: Filter and limit
- **WHEN** a configuration sets `filter_name = "eu"` and `max_items = 10`
- **THEN** each request carries `filter[name]=eu`, and fetching stops once 10 items are read

### Requirement: List paging
Lists paged with `page[number]` or `page_number` SHALL stop at `meta.total_pages` or a short page. Lists paged only with `page[after]` SHALL follow `meta.cursors.after` or `meta.end_cursor` until it is empty. When an operation offers both numbers and cursors, page numbers SHALL be used, and other paging parameters SHALL be ignored.

#### Scenario: Paged list
- **WHEN** `data "telnyx_billing_groups" "all" {}` is read and the API returns three pages with `meta.total_pages = 3`
- **THEN** `billing_groups` holds the items of all three pages in API order

#### Scenario: Cursor-paged list
- **WHEN** `data "telnyx_pay_connectors" "all" {}` is read and the first page's `meta.end_cursor` is `c1`, the second's is empty
- **THEN** the second request carries `page[after]=c1`, no third request is sent, and the items of both pages are returned

#### Scenario: Unsupported paging
- **WHEN** a list operation pages only with parameters the engine does not drive, such as `offset`, and the override does not set `paging: none`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Record paging
Lists paged with a plain `page` and `recordsPerPage`, or `page` and `page_size` without `page_number`, SHALL send both. When the response gives the item count (`totalRecords` or `total_records`) they SHALL page until that many items are read or a page is empty; otherwise they SHALL stop at a short page.

#### Scenario: Record-paged list
- **WHEN** `data "telnyx_10dlc_brands" "all" {}` is read and the API answers `{"records": [...], "totalRecords": 4}` with two records per page
- **THEN** the requests carry `page=1` and `page=2` with `recordsPerPage`, no third request is sent, and `brands` holds the four records

### Requirement: Data-source-only overrides
An override with `kind: data_source` SHALL produce a data source with no resource. It reads either one object (`read`) or a list (`list`). Path parameters SHALL be Required lookup attributes. A `read` operation's scalar string query parameters SHALL be lookup attributes too: Required when the spec requires them, otherwise Optional, and sent only when set.

#### Scenario: Reference data
- **WHEN** an override declares `kind: data_source` with `read: GET /balance`
- **THEN** `data "telnyx_balance"` exists with every attribute Computed, and there is no `telnyx_balance` resource

#### Scenario: Lookup by query
- **WHEN** `data "telnyx_sip_registration_status" "x" { credential_type = "telephony_credential", connection_id = "c1" }` is read
- **THEN** the provider sends `GET /sip_registration_status?connection_id=c1&credential_type=telephony_credential` and `username` is not sent

### Requirement: Read from a collection
A resource whose override sets `read_from` SHALL read itself by finding the element whose `key` field (or whose value, for plain values) equals its id. It searches the list operation's pages, or a field of its parent object. When no element matches, the resource SHALL be treated as deleted outside Terraform.

#### Scenario: Read from the list
- **WHEN** `telnyx_integration_secret` refreshes and `GET /integration_secrets` lists its id on page 2
- **THEN** state is filled from that element and the plan is empty

#### Scenario: Gone from the list
- **WHEN** the id is on no page
- **THEN** the resource is removed from state, as for a 404

#### Scenario: Plain value in the parent
- **WHEN** `telnyx_ai_assistant_tag` with tag `vip` refreshes and the assistant's `tags` contains `vip`
- **THEN** the resource stays in state, and when `tags` no longer contains it, the resource is removed from state

### Requirement: Extra update endpoints
A resource override MAY list `updates`, each an operation that sets some attributes, with an optional `clear` operation. The attributes of an update operation SHALL be configurable without forcing replacement. The engine SHALL call an update operation after create when its attributes are set, and on update only when they changed. It SHALL call `clear` instead when they were all removed.

#### Scenario: Change one setting
- **WHEN** only `wireless_blocklist_id` of a `telnyx_sim_card_group` changes
- **THEN** the engine calls only `POST /sim_card_groups/{id}/actions/set_wireless_blocklist`, and the next plan is empty

#### Scenario: Remove a setting
- **WHEN** `wireless_blocklist_id` is removed from configuration
- **THEN** the engine calls `POST /sim_card_groups/{id}/actions/remove_wireless_blocklist`

### Requirement: Actions
An override with `kind: action` SHALL produce a Terraform action whose configuration holds the operation's path parameters (Required) and request body fields. Invoking it SHALL send the operation once with the fields that are set, and an API error SHALL fail the action.

#### Scenario: Invoke an action
- **WHEN** `action "telnyx_ai_assistant_clone" "c" { config { assistant_id = "a1" } }` is invoked
- **THEN** the provider sends `POST /ai/assistants/a1/clone` once and stores nothing in state

#### Scenario: Action fails
- **WHEN** the API answers the action's request with an error
- **THEN** the action fails with the API's error message

### Requirement: Ephemeral resources
An override with `kind: ephemeral` SHALL produce an ephemeral resource. Its path parameters SHALL be Required and its response fields Computed and Sensitive. The value SHALL never be written to plan or state.

#### Scenario: Read a secret
- **WHEN** `ephemeral "telnyx_sim_card_activation_code" "x" { id = "sim1" }` is opened
- **THEN** its attributes hold the response of `GET /sim_cards/sim1/activation_code`, and neither plan nor state contains them

### Requirement: Write-only secrets
An attribute override with `write_only: true` SHALL produce a write-only, sensitive attribute named `<name>_wo` and an optional `<name>_wo_version`. Create SHALL send the write-only value. Update SHALL send it only when `<name>_wo_version` changed. The value SHALL never be stored in state.

#### Scenario: Create with a secret
- **WHEN** a resource is created with `password_wo = "s3cret"`
- **THEN** the create request carries `password`, and state holds no `password_wo` value

#### Scenario: Rotate a secret
- **WHEN** `password_wo` changes but `password_wo_version` does not
- **THEN** the plan is empty
- **WHEN** `password_wo_version` changes as well
- **THEN** the update request carries the new `password`

### Requirement: Release on destroy
An override with `release_on_destroy: true` SHALL add an optional `release_on_destroy` attribute, false by default and never sent to the API. Destroy SHALL call the delete operation only when it is true in state. Otherwise destroy SHALL remove the object from state and warn that it still exists in Telnyx.

#### Scenario: Keep by default
- **WHEN** a phone number with no `release_on_destroy` is destroyed
- **THEN** no delete request is sent, the number leaves state, and a warning says it is still in Telnyx

#### Scenario: Release
- **WHEN** a phone number with `release_on_destroy = true` in state is destroyed
- **THEN** the delete operation is sent

### Requirement: Settings objects
An override with `id_from_path` SHALL identify the object by that path parameter. Every path parameter of the create path SHALL be a Required attribute that forces replacement, and import SHALL take them all in path order, separated by `/`.

#### Scenario: Import nested settings
- **WHEN** a user runs `terraform import telnyx_external_connection_phone_number_settings.x ec1/pn1`
- **THEN** `external_connection_id` is `ec1`, `phone_number_id` is `pn1`, and the settings are read from `GET /external_connections/ec1/phone_numbers/pn1`

### Requirement: Once-per-account settings
An override with `singleton: true` SHALL produce a resource whose operations have no path parameters. Its `id` SHALL always be the type name without `telnyx_`, and its data source SHALL take no lookup attribute. Destroy SHALL follow the rule for objects without delete.

#### Scenario: Manage account settings
- **WHEN** `resource "telnyx_inbound_voice_channel_settings" "x" { channels = 10 }` is applied
- **THEN** the provider sends `PATCH /inbound_channels` with `channels`, `id` is `inbound_voice_channel_settings`, and the next plan is empty

### Requirement: Plain-text ephemeral values
When an ephemeral resource's operation answers in `text/plain`, the ephemeral resource SHALL have one Computed, Sensitive attribute holding the text without surrounding whitespace. Overrides MAY rename it.

#### Scenario: Mint a token
- **WHEN** `ephemeral "telnyx_telephony_credential_token" "t" { id = "c1" }` is opened
- **THEN** `token` holds the body of `POST /telephony_credentials/c1/token`, and neither plan nor state contains it

### Requirement: User-named objects
When an override's `id` names a response field that the create request also takes, such as a connector's `name` or a verified `phone_number`, that attribute SHALL be Required, sent on create, force replacement, and be the import id.

#### Scenario: Create a SIPREC connector
- **WHEN** `telnyx_siprec_connector` is created with `name = "rec"`
- **THEN** the create request carries `name`, later requests use `/siprec_connectors/rec`, and import takes `rec`

### Requirement: Masked secrets
An attribute override with `masked: true` SHALL make the attribute sensitive and SHALL keep its value in state when the API answers with only asterisks, as Telnyx does for passwords after create.

#### Scenario: Password masked on read
- **WHEN** a credential connection's password is `s3cret` in state and the API returns `********`
- **THEN** the plan is empty
- **WHEN** the API returns another value
- **THEN** the plan updates the password to the configured value

### Requirement: Secrets in lists
A resource's sensitive, masked and write-only attributes SHALL be sensitive in the items of its list data source too, and its sets SHALL be sets there.

#### Scenario: List credentials
- **WHEN** `data "telnyx_telephony_credentials" "all" {}` is read
- **THEN** each item's `sip_password` is sensitive

### Requirement: Form request bodies
When an operation takes only `application/x-www-form-urlencoded`, the provider SHALL send its scalar fields as a form. A request body of a type the provider cannot send SHALL be a generation error.

#### Scenario: Create a TeXML queue
- **WHEN** `telnyx_texml_queue` is created with `max_size = 5`
- **THEN** the request is a form with `MaxSize=5`

### Requirement: Untyped string enums
A schema with an `enum` of strings, or a string `const`, and no `type` SHALL be a string; a `const` SHALL be its only allowed value. When an `allOf` part narrows a property that an earlier part defines, the property SHALL keep the earlier definition with the narrowing applied.

#### Scenario: Feature list
- **WHEN** an advanced number order sets `features = ["sms", "voice"]`
- **THEN** the values are strings, not JSON

#### Scenario: Narrowed contact type
- **WHEN** an RCS brand's `contacts.brand` is a `BrandContact` whose `contact_type` an `allOf` part fixes to `BRAND`
- **THEN** `contact_type` is a string whose only allowed value is `BRAND`

### Requirement: Echoed request objects
When the response gives an object no fields, or gives a map whose entries have no single shape, the object's configured fields SHALL be read back from the response by the request's schema, so import and drift detection cover them. A map whose entries have fields SHALL be read by the entry schema.

#### Scenario: Import of named identifiers
- **WHEN** an RCS brand with `identifiers.ein` is imported and the API returns `identifiers` as a map of identifiers
- **THEN** `identifiers.ein.value` is in state

### Requirement: Names starting with a digit
Type names whose first segment starts with a digit, such as `telnyx_10dlc_brand`, SHALL generate valid Go names, and a list data source's items attribute SHALL drop that segment (`telnyx_10dlc_brands` lists `brands`).

#### Scenario: 10DLC list
- **WHEN** `data "telnyx_10dlc_campaigns" "all" {}` is read
- **THEN** the campaigns are in `campaigns`

### Requirement: Free-form read responses
A data source whose read response is a map without fixed fields, or has no fields the generator can derive (such as a union of lists and objects), SHALL expose the whole response as one computed JSON attribute named `values`.

#### Scenario: Campaign operation status
- **WHEN** `data "telnyx_10dlc_campaign_operation_status" "x" { campaign_id = "c1" }` is read and the API answers with one entry per carrier
- **THEN** `values` holds that map as JSON

### Requirement: Versioned spec paths
A spec path that starts with `/v2/` SHALL be sent once under the base URL's `/v2`, not as `/v2/v2/...`.

#### Scenario: Mobile phone numbers
- **WHEN** `telnyx_mobile_phone_number` reads `GET /v2/mobile_phone_numbers/{id}` with the default base URL
- **THEN** the request goes to `https://api.telnyx.com/v2/mobile_phone_numbers/{id}`

### Requirement: Attribute names from camelCase
JSON field names SHALL become snake_case attribute names, with a run of capitals as one word: `webhookURL` is `webhook_url` and `isTMobileRegistered` is `is_t_mobile_registered`. Path parameter renames (`path_params`) SHALL apply to list data sources too (see "Parent names in list data sources").

#### Scenario: Webhook URL
- **WHEN** a 10DLC brand's `webhookURL` is derived
- **THEN** the attribute is `webhook_url`, on the resource and on `telnyx_10dlc_brands`

### Requirement: Fields with different request and response shapes
A field the API takes in one JSON type and returns in another (a boolean returned as a string, strings returned as objects) SHALL be request-only: it is sent but not read back, unless the override sets `returned` because a hook reshapes it.

#### Scenario: Hosted number order numbers
- **WHEN** a hosted number order is created with `phone_numbers = ["+15551230000"]` and the API returns the numbers as objects
- **THEN** create succeeds, and `phone_numbers` keeps the configured strings

### Requirement: Values only create returns
An attribute override MAY set `create_only: true` for a top-level response field that no request takes. The attribute SHALL be computed and sensitive. When the read response schema lacks the field, the generator SHALL take it from the create response schemas, looking at every 2xx response from the lowest code up, and SHALL fail when none lists it. The engine SHALL fill the value from the create response and SHALL keep the value in state on every read.

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

### Requirement: Invalid create-only fields
Setting `create_only` on a field a request takes, or on a nested field, SHALL be a generation error.

#### Scenario: Configurable field
- **WHEN** an override sets `create_only` on `name`, which the create request takes
- **THEN** generation fails naming the field

### Requirement: Create-only values in data sources
Data sources and list data sources SHALL leave create-only attributes out, because a read never returns them.

#### Scenario: Tunnel data source
- **WHEN** `data "telnyx_tunnel"` is read
- **THEN** it has no `tunnel_token` attribute

### Requirement: Read-only request objects
A request object whose fields are all `readOnly` SHALL not be configurable: the attribute SHALL be computed from the response. An object with at least one writable field SHALL stay configurable.

#### Scenario: SIM card status
- **WHEN** `PATCH /sim_cards/{id}` lists `status` with only read-only fields
- **THEN** `telnyx_sim_card.status` is computed and cannot be set in configuration

### Requirement: Objects identified by several fields
A `read_from` override MAY list `match`: further attributes whose values the collection element's fields SHALL equal, besides the key. Match attributes SHALL be required and force replacement, and the import id SHALL be the id followed by them, separated by `/`. Generation SHALL fail when a match attribute does not exist.

#### Scenario: Limit per product and period
- **WHEN** `GET /spend_limits` lists `voice` for `daily` and `monthly`, and the state has `product = "voice"`, `period = "monthly"`
- **THEN** the monthly entry is read

#### Scenario: Import
- **WHEN** `telnyx_spend_limit` is imported as `voice/daily`
- **THEN** `product` is `voice` and `period` is `daily`

### Requirement: Query parameters on update and delete
An override MAY list `query`: attributes that update and delete operations declaring a query parameter of the same name SHALL receive as that query parameter when set. Generation SHALL fail when the attribute does not exist or no update or delete operation takes it.

#### Scenario: Period of a limit
- **WHEN** the amount of a monthly `voice` limit changes
- **THEN** the engine sends `PATCH /spend_limits/voice?period=monthly`

### Requirement: Unions told apart by a boolean
When the variants of a `oneOf` or `anyOf` object have no `type` tag but share a boolean property with one different value in each, the generator SHALL merge them into one object: the boolean without its enum, the union of the properties, and as required only the fields every variant requires.

#### Scenario: Amount or unlimited
- **WHEN** `POST /spend_limits` takes an amount variant (`unlimited: false`) or an unlimited variant (`unlimited: true`)
- **THEN** `telnyx_spend_limit` has optional `amount` and optional boolean `unlimited`

### Requirement: Collection operations
An operation whose path parameters are exactly those of the create path SHALL fill them from the parent attributes, not from the object's id.

#### Scenario: Delete by body
- **WHEN** a DIR phone number is deleted with `DELETE /dir/{dir_id}/phone_numbers`
- **THEN** `dir_id` comes from the `dir_id` attribute and the number is named by the delete hook's body

### Requirement: Objects named by a response field
When the override's `id` names a response field that the override makes required, the field SHALL name the object as a create field does: required, forcing replacement, and filled into the paths that take it. A hook sends it.

#### Scenario: Reputation number
- **WHEN** `telnyx_enterprise_reputation_number` sets `phone_number = "+13125550100"`
- **THEN** the create hook sends `{"phone_numbers": ["+13125550100"]}` and the read path is `/enterprises/{enterprise_id}/reputation/numbers/+13125550100`

### Requirement: Once-per-account objects with an id
A `singleton` whose response has an `id` field SHALL use the fixed id in its place.

#### Scenario: Auto recharge
- **WHEN** `GET /payment/auto_recharge_prefs` returns `id: "123"`
- **THEN** the `id` of `telnyx_auto_recharge_preferences` is `auto_recharge_preferences`

### Requirement: Delete hooks and gone objects
A resource's `Delete` hook SHALL replace the generic delete call, with not-found errors ignored. A not-found error from a resource's read hook SHALL remove the object from state, as a 404 from the read does.

#### Scenario: Limit removed outside Terraform
- **WHEN** the spend limit entry of a product and period has `limit: null`
- **THEN** the next plan creates the limit again

### Requirement: Skipped query lookups
A data-source-only `read` override MAY set `attributes.<param>.skip: true` for a query parameter; the parameter SHALL then not become a lookup attribute and SHALL never be sent.

#### Scenario: Billed refresh
- **WHEN** `telnyx_account_reputation_number` skips `fresh`
- **THEN** the data source has no `fresh` attribute and reads cached reputation data

### Requirement: Values not returned after import
After import, when a value that forces replacement but that the API does not return changes from null, the provider SHALL update the state without replacing the object and without sending a request for it.

#### Scenario: Imported managed account
- **WHEN** a `telnyx_managed_account` is imported and the configuration sets `business_name`
- **THEN** the plan updates in place, and no second managed account is created

### Requirement: Escaped path parameters
Path parameter values SHALL be percent-encoded, with `+` sent as `%2B`, and the client SHALL send the escaped path without escaping it again.

#### Scenario: Phone number in a path
- **WHEN** `telnyx_enterprise_reputation_number` for `+13125550100` is read
- **THEN** the request path ends in `/reputation/numbers/%2B13125550100`

### Requirement: Extra updates without a clear
Attributes of an update without `clear` that the API returns SHALL be optional and computed unless create requires them. After create, an update operation SHALL NOT be called for attributes the create request sent.

#### Scenario: Server default without a clear
- **WHEN** a `telnyx_enterprise_reputation` is created without `check_frequency` and the API answers `business_daily`
- **THEN** `check_frequency` is `business_daily` in state and the next plan is empty

### Requirement: Raw bodies
A request body of `application/octet-stream` alone SHALL NOT be a generation error; the operation SHALL be marked raw, and sending it without a `Send` hook SHALL fail. A read that answers with plain text or raw bytes SHALL be read as one `value` attribute holding the body unchanged.

#### Scenario: KV value drift
- **WHEN** a KV key's value is changed outside Terraform
- **THEN** the next plan updates `value`

### Requirement: Lists in results
A list response that holds its items in `results` and their total in `count`, paged with `page` and `page_size`, SHALL be read page by page like lists in `records`: until `count` items are read or a page is empty, and without a count until a short page.

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

### Requirement: Forced deletes
An override's `force_destroy` SHALL name query parameters that take a boolean (alone or as a variant of a union) of the resource's delete operation, and SHALL add a local, optional boolean attribute `force_destroy`. When `force_destroy` is true in state, destroy SHALL send each named parameter as `true`; otherwise it SHALL send none. Generation SHALL fail when the override has no delete, or a name is not a boolean query parameter of the delete operation.

#### Scenario: Force a delete
- **WHEN** a resource with `force_destroy: [force, drain]` and `force_destroy = true` in state is destroyed
- **THEN** the delete request carries `drain=true&force=true`

#### Scenario: Plain delete
- **WHEN** the same resource without `force_destroy` is destroyed
- **THEN** the delete request carries no query parameters

#### Scenario: Unknown parameter
- **WHEN** an override names `force_destroy: [purge]` and the delete operation has no boolean query parameter `purge`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Union list responses
When a list operation's response has no `data` array but is a `oneOf` or `anyOf` of objects, the generator SHALL take the items from the first variant whose `data` is an array. Paging SHALL be derived from the operation's parameters as for other lists.

#### Scenario: Two page shapes
- **WHEN** `GET /email_blocks` answers with `oneOf` a page-number shape and a cursor shape
- **THEN** `data "telnyx_email_suppressions"` lists the items, paged by `page[number]`

### Requirement: Lowercase values
An attribute override with `lowercase: true` SHALL reject a configured value that has capital letters or leading or trailing spaces; for a list or set of strings, each element. The error SHALL show at plan time.

#### Scenario: Mixed case address
- **WHEN** a configuration sets `to = "Someone@Example.com"` on an attribute with `lowercase: true`
- **THEN** plan fails with an error saying the value must be lowercase

### Requirement: Soft-deleted objects
Objects the API soft-deletes SHALL have a read hook that answers with a not-found error when the object's status says it is deleted, so read removes it from state (see "Delete hooks and gone objects").

#### Scenario: Removed suppression
- **WHEN** `GET /email_blocks/{id}` answers with `status = "removed"`
- **THEN** the suppression leaves state and the next plan creates it again

### Requirement: Local changes without an update
When a resource has no update operation and only local attributes (such as `force_destroy`) change, apply SHALL store the new values without a request.

#### Scenario: Set force_destroy on a membership
- **WHEN** `force_destroy = true` is added to an existing `telnyx_email_ip_pool_ip`
- **THEN** the plan updates it in place, apply sends no request, and destroy then sends `force=true&drain=true`

### Requirement: Reserved filter names
A list data source's query parameter whose snake_case name is reserved by Terraform (a meta-argument such as `provider` or `count`) or by the list itself (`max_items`) SHALL become the filter attribute `<name>_filter`, sent under the parameter's own name. A path parameter with a reserved name, or a renamed filter that collides with another parameter, SHALL be a generation error.

#### Scenario: Provider filter
- **WHEN** `data "telnyx_speech_to_text_providers" "x" { provider_filter = "deepgram" }` is read
- **THEN** the request is `GET /speech-to-text/providers?provider=deepgram`

#### Scenario: Collision
- **WHEN** an operation takes both `provider` and `provider_filter` as query parameters
- **THEN** generation fails naming both parameters

### Requirement: Nullable action fields
An attribute override with `nullable: true` on a top-level field of an action request SHALL make the attribute optional; invoking the action with the attribute unset SHALL send the field as JSON `null`, and with a value SHALL send the value. `nullable` on any other attribute SHALL be a generation error.

#### Scenario: Remove a number from its bundle
- **WHEN** `action "telnyx_phone_number_bundle_status_change" "x" { config { id = "pn1" } }` is invoked
- **THEN** the provider sends `PATCH /phone_numbers/pn1/actions/bundle_status_change` once with `{"bundle_id":null}`

#### Scenario: Nullable outside an action
- **WHEN** a data source override sets `nullable: true` on an attribute
- **THEN** generation fails naming the attribute
