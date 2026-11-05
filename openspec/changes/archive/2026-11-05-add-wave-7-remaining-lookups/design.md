# Design

Every object in this wave is generated from an override without hooks. None is a resource: the coverage review classified each operation as a read-only data source or a one-shot action.

## Read or list

An override with `read` makes a data source for one object; `list` makes a list data source with filters, paging and `max_items`. The generator decides which fits from the response: a `data` array is a list. Some reads have no path parameter (`GET /balance`, `GET /connections/count`, `GET /charges_summary`, `GET /terms_of_service/status`): they are single-object data sources whose query parameters are lookup attributes, so `telnyx_charges_summary` requires `start_date` and `end_date`.

Single-object reads take only scalar query parameters. `GET /user_tags` also takes `filter[starts_with]`, which the data source does not offer; users filter the returned tags in configuration. `GET /charges_breakdown` takes `format`; `csv` returns a file the provider cannot read, which the description says.

## Reserved filter names

`GET /speech-to-text/providers` filters by `provider`, which Terraform reserves in every block. A list filter whose name is reserved now takes the suffix `_filter` (`provider_filter`) and is still sent as `provider`. If the operation already has a parameter of that name, generation fails as for any other collision. Path parameters with a reserved name stay a generation error, because they also form import ids.

## Secrets

`external_pin` in `telnyx_phone_numbers_bulk_update` is sensitive, as it is on `telnyx_phone_number`. A port-out's `support_key` (`telnyx_portout`, `telnyx_portouts`) proves the account holder's request, so it is sensitive too.

## Actions

The actions send their operation once, as the actions from step 2 do. Path parameters are required attributes. `telnyx_portout_status_update` takes the new status as the `status` path parameter (`authorized` or `rejected-pending`); the generator does not carry path-parameter enums, so the description lists the values and the API rejects others. The API requires `reason` in both cases. Destructive one-shot operations (`telnyx_phone_numbers_bulk_delete`, `telnyx_organization_user_remove`, `telnyx_ai_integration_connection_delete`, `telnyx_user_bundle_deactivate`) are actions because the objects they remove are not managed by this provider.

## Not built here

- `GET /bulk_sim_card_actions`, `GET /bulk_sim_card_actions/{id}` and `GET /wireless/regions`: wave 3, which owns SIM cards and wireless.
- `GET /telephony_credentials/{id}/token_material`: wave 6.
- `GET /organizations/users/{id}` (boolean `include_groups`) and `GET /requirements/{id}` (integer `version`): wave 5 adds boolean and integer lookup parameters to the generator. The two overrides follow in a later change once it merges; until then the coverage entries stay `data_source` and unimplemented.

## Testing

Unit tests cover the reserved-name rule. The conformance test covers resources only; the new data sources and actions are checked by `go run ./cmd/codegen` and the provider's schema validation in `go test ./...`. Acceptance tests are not run in this change (no API key in the build environment).
