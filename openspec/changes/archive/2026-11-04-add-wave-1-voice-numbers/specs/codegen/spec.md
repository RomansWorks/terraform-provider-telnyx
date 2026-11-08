## MODIFIED Requirements

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

### Requirement: List data sources
A resource override's `list` operation, or a data-source-only override, SHALL produce a data source with:
- optional filter attributes derived from the operation's query parameters, excluding paging
- an optional `max_items`
- one computed list attribute of the items

The engine SHALL fetch pages until the last page or `max_items`, and SHALL send only the filters that are set. Lists paged with `page[number]` or `page_number` SHALL stop at `meta.total_pages` or a short page. Lists paged only with `page[after]` SHALL follow `meta.cursors.after` or `meta.end_cursor` until it is empty. When an operation offers both, page numbers SHALL be used. Once a supported style is found, other paging parameters of the operation SHALL be ignored.

#### Scenario: Paged list
- **WHEN** `data "telnyx_billing_groups" "all" {}` is read and the API returns three pages with `meta.total_pages = 3`
- **THEN** `billing_groups` holds the items of all three pages in API order

#### Scenario: Cursor-paged list
- **WHEN** `data "telnyx_pay_connectors" "all" {}` is read and the first page's `meta.end_cursor` is `c1`, the second's is empty
- **THEN** the second request carries `page[after]=c1`, no third request is sent, and the items of both pages are returned

#### Scenario: Filter and limit
- **WHEN** a configuration sets `filter_name = "eu"` and `max_items = 10`
- **THEN** each request carries `filter[name]=eu`, and fetching stops once 10 items are read

#### Scenario: Unsupported paging
- **WHEN** a list operation pages only with parameters the engine does not drive, such as `offset`, and the override does not set `paging: none`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Data-source-only overrides
An override with `kind: data_source` SHALL produce a data source with no resource. It reads either one object (`read`) or a list (`list`). Path parameters SHALL be Required lookup attributes. A `read` operation's scalar string query parameters SHALL be lookup attributes too: Required when the spec requires them, otherwise Optional, and sent only when set.

#### Scenario: Reference data
- **WHEN** an override declares `kind: data_source` with `read: GET /balance`
- **THEN** `data "telnyx_balance"` exists with every attribute Computed, and there is no `telnyx_balance` resource

#### Scenario: Lookup by query
- **WHEN** `data "telnyx_sip_registration_status" "x" { credential_type = "telephony_credential", connection_id = "c1" }` is read
- **THEN** the provider sends `GET /sip_registration_status?connection_id=c1&credential_type=telephony_credential` and `username` is not sent

## ADDED Requirements

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
A schema with an `enum` of strings and no `type` SHALL be a string.

#### Scenario: Feature list
- **WHEN** an advanced number order sets `features = ["sms", "voice"]`
- **THEN** the values are strings, not JSON
