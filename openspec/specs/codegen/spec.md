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
When a response schema's only property is `data`, the provider SHALL read the object from `data`.

#### Scenario: Wrapped response
- **WHEN** the API returns `{"data": {...}}`
- **THEN** the resource's attributes come from the inner object

### Requirement: Shared lifecycle behaviour
Every generated resource SHALL remove itself from state when read returns 404, support import by id (or `parent_id/id` for nested paths), and send only changed attributes on PATCH or POST updates and the full object on PUT updates.

#### Scenario: Deleted outside Terraform
- **WHEN** an object was deleted in the Telnyx portal
- **THEN** the next plan proposes to create it again

#### Scenario: Import
- **WHEN** a user runs `terraform import telnyx_ai_assistant.x <id>`
- **THEN** the resource is read into state and the next plan is empty for a matching configuration

### Requirement: Objects without delete
When the API offers no delete for an object, destroy SHALL remove it from state and warn that the object still exists in Telnyx.

#### Scenario: Destroy without delete
- **WHEN** a resource with no delete operation is destroyed
- **THEN** no API call is made, the object leaves state, and a warning names it

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

#### Scenario: Paged list
- **WHEN** `data "telnyx_billing_groups" "all" {}` is read and the API returns three pages with `meta.total_pages = 3`
- **THEN** `billing_groups` holds the items of all three pages in API order

#### Scenario: Filter and limit
- **WHEN** a configuration sets `filter_name = "eu"` and `max_items = 10`
- **THEN** each request carries `filter[name]=eu`, and fetching stops once 10 items are read

#### Scenario: Unsupported paging
- **WHEN** a list operation pages with a parameter the engine does not know and the override does not set `paging: none`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Data-source-only overrides
An override with `kind: data_source` SHALL produce a data source with no resource. It reads either one object (`read`) or a list (`list`). Path parameters SHALL be Required lookup attributes.

#### Scenario: Reference data
- **WHEN** an override declares `kind: data_source` with `read: GET /balance`
- **THEN** `data "telnyx_balance"` exists with every attribute Computed, and there is no `telnyx_balance` resource

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
