## ADDED Requirements

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
