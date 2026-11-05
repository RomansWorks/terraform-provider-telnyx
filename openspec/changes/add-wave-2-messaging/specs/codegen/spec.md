## MODIFIED Requirements

### Requirement: List paging
Lists paged with `page[number]` or `page_number` SHALL stop at `meta.total_pages` or a short page. Lists paged only with `page[after]` SHALL follow `meta.cursors.after` or `meta.end_cursor` until it is empty. Lists paged with a plain `page` and `recordsPerPage`, or `page` and `page_size` without `page_number`, SHALL send both and stop at a short page or once `totalRecords` items are read. When an operation offers both numbers and cursors, page numbers SHALL be used, and other paging parameters SHALL be ignored.

#### Scenario: Paged list
- **WHEN** `data "telnyx_billing_groups" "all" {}` is read and the API returns three pages with `meta.total_pages = 3`
- **THEN** `billing_groups` holds the items of all three pages in API order

#### Scenario: Cursor-paged list
- **WHEN** `data "telnyx_pay_connectors" "all" {}` is read and the first page's `meta.end_cursor` is `c1`, the second's is empty
- **THEN** the second request carries `page[after]=c1`, no third request is sent, and the items of both pages are returned

#### Scenario: Record-paged list
- **WHEN** `data "telnyx_10dlc_brands" "all" {}` is read and the API answers `{"records": [...], "totalRecords": 4}` with two records per page
- **THEN** the requests carry `page=1` and `page=2` with `recordsPerPage`, no third request is sent, and `brands` holds the four records

#### Scenario: Unsupported paging
- **WHEN** a list operation pages only with parameters the engine does not drive, such as `offset`, and the override does not set `paging: none`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Untyped string enums
A schema with an `enum` of strings, or a string `const`, and no `type` SHALL be a string; a `const` SHALL be its only allowed value. When an `allOf` part narrows a property that an earlier part defines, the property SHALL keep the earlier definition with the narrowing applied.

#### Scenario: Feature list
- **WHEN** an advanced number order sets `features = ["sms", "voice"]`
- **THEN** the values are strings, not JSON

#### Scenario: Narrowed contact type
- **WHEN** an RCS brand's `contacts.brand` is a `BrandContact` whose `contact_type` an `allOf` part fixes to `BRAND`
- **THEN** `contact_type` is a string whose only allowed value is `BRAND`

## ADDED Requirements

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
A data source whose read response is a map without fixed fields SHALL expose the whole response as one computed JSON attribute named `values`.

#### Scenario: Campaign operation status
- **WHEN** `data "telnyx_10dlc_campaign_operation_status" "x" { campaign_id = "c1" }` is read and the API answers with one entry per carrier
- **THEN** `values` holds that map as JSON

### Requirement: Versioned spec paths
A spec path that starts with `/v2/` SHALL be sent once under the base URL's `/v2`, not as `/v2/v2/...`.

#### Scenario: Mobile phone numbers
- **WHEN** `telnyx_mobile_phone_number` reads `GET /v2/mobile_phone_numbers/{id}` with the default base URL
- **THEN** the request goes to `https://api.telnyx.com/v2/mobile_phone_numbers/{id}`

### Requirement: Attribute names from camelCase
JSON field names SHALL become snake_case attribute names, with a run of capitals as one word: `webhookURL` is `webhook_url` and `isTMobileRegistered` is `is_t_mobile_registered`. Path parameter renames (`path_params`) SHALL apply to list data sources too.

#### Scenario: Webhook URL
- **WHEN** a 10DLC brand's `webhookURL` is derived
- **THEN** the attribute is `webhook_url`, on the resource and on `telnyx_10dlc_brands`
