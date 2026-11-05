## MODIFIED Requirements

### Requirement: Extra update endpoints
A resource override MAY list `updates`, each an operation that sets some attributes, with an optional `clear` operation. The attributes of an update operation SHALL be configurable without forcing replacement. The engine SHALL call an update operation after create when its attributes are set, and on update only when they changed. It SHALL call `clear` instead when they were all removed. Attributes of an update without `clear` that the API returns SHALL be optional and computed, so that a server default does not show as a change, unless create requires them. After create, an update operation SHALL NOT be called for attributes the create request sent.

#### Scenario: Change one setting
- **WHEN** only `wireless_blocklist_id` of a `telnyx_sim_card_group` changes
- **THEN** the engine calls only `POST /sim_card_groups/{id}/actions/set_wireless_blocklist`, and the next plan is empty

#### Scenario: Remove a setting
- **WHEN** `wireless_blocklist_id` is removed from configuration
- **THEN** the engine calls `POST /sim_card_groups/{id}/actions/remove_wireless_blocklist`

#### Scenario: Server default without a clear
- **WHEN** a `telnyx_enterprise_reputation` is created without `check_frequency` and the API answers `business_daily`
- **THEN** `check_frequency` is `business_daily` in state and the next plan is empty

## ADDED Requirements

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

### Requirement: Values returned once
`attributes.<f>.once: true` SHALL make a response field sensitive and computed, take its value from the create response, and keep the value in state when later reads return none. Data sources SHALL show what the API returns.

#### Scenario: Client secret
- **WHEN** a confidential `telnyx_oauth_client` is created and `GET /oauth/clients/{id}` returns `client_secret: null`
- **THEN** `client_secret` in state holds the secret from the create response, and the next plan is empty

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
