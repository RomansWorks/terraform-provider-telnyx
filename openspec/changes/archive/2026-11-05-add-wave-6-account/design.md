# Design

Wave 6 is mostly overrides. The generator learned the shapes it could not derive, each as a general rule, and no existing descriptor changes.

## Spend limits

```yaml
type: telnyx_spend_limit
create: POST /spend_limits
update: PATCH /spend_limits/{product}
delete: DELETE /spend_limits/{product}
id: product
read_from: {op: "GET /spend_limits", key: product, match: [period]}
query: [period]
```

A limit exists per product and period, and `GET /spend_limits` lists one entry per pair. `read_from.match` adds `period` to the lookup: the element must match the id and every match attribute. Match attributes are required, force replacement and follow the id in the import id (`voice/daily`). The update and delete paths name only the product; the period is a query parameter (it defaults to `daily` on the server). `query` makes the engine send the attribute as a query parameter on every update and delete operation that declares it; it is an error when none does.

The create body is `oneOf` an amount variant (`unlimited: false`) and an unlimited variant (`unlimited: true`). Variants told apart by a boolean with one value each now merge into one object: the flag keeps its type without the enum, and a field is required only when every variant requires it. Variants without a `type` or boolean tag still become JSON. A hook sends exactly one of `amount` and `unlimited` on every request, also when only `reason` changes, because the API replaces the limit.

The list also has an entry for every product and period without a limit (`limit: null`). The read hook reports such an entry as not found, so a limit removed outside Terraform is created again. Read hooks may now return a not-found error for this; before, only the read call could.

API key limits (`credential_id`) are left out: the list nests them in `credential_limits` of the product entry, which `read_from` cannot address. They can follow as their own resource.

## Batches of phone numbers

`POST /dir/{dir_id}/phone_numbers` and `POST /enterprises/{id}/reputation/numbers` take `phone_numbers` and answer with a list. Each Terraform resource is one number: the override skips `phone_numbers`, makes the response field `phone_number` required, and a hook sends it as a list of one. The create response is read from its only element, as for porting orders.

A response field the override makes required now names the object like a create field does (required, forces replacement), so `telnyx_enterprise_reputation_number` is identified by `phone_number`, which its read and delete paths take.

`DELETE /dir/{dir_id}/phone_numbers` names the numbers in a body. Two changes make that work:
- An operation whose path parameters are exactly the create path's is an operation on the collection: all its parameters are parents. Before, the last one was taken as the object's id.
- `Hooks.Delete` replaces the generic delete call; the DIR phone number hook sends `{"phone_numbers": [number]}`. A not-found error is ignored as for the generic delete.

## Once-per-account objects with their own id

`GET /payment/auto_recharge_prefs` and `GET /user_profile` return an `id` (the user profile's is an integer). It identifies nothing the user chooses, so the fixed id replaces it instead of failing generation.

## Values returned once

`client_secret` of an OAuth client is returned by create only. It uses `create_only` from wave 3, which was merged first; this change's own `once` key was dropped in favour of it.

## Extra updates without a clear

Attributes set by an extra update are optional and never computed, so that removing one runs the clear operation. When there is no clear operation, nothing could run, and a server default (the reputation `check_frequency`) would show as a change forever. Those attributes are now optional and computed: the API's value is kept until the user sets one.

## Review fixes

- **Values the API does not return, after import.** A managed account's `business_name` and `password`, or a DIR phone number's `documents`, are sent on create and never returned, and they force replacement. After import they are null, so setting them in configuration would replace the object, which for a managed account means a second account the API cannot delete. Import now marks the state as imported (private state), and a change from null of a value the API does not return then updates the state instead of replacing. A resource without update operations accepts such an update without sending anything.
- **Phone numbers in paths.** `url.PathEscape` keeps `+`, which some servers read as a space; Telnyx asks for `%2B`. Path parameters now escape `+`, and the client keeps an escaped path as it is instead of escaping `%` again.
- **Reputation LOA.** Replacing the LOA document is an update of `telnyx_enterprise_reputation` (`PATCH /enterprises/{id}/reputation/loa`), not an action: as an action, the next plan would have destroyed monitoring to restore the old document id. An extra update now skips attributes the create request already sent, and an attribute create requires stays required.
- **Spend limit drift.** The read hook copies `limit.amount` (and `limit.unlimited` when true) to the top level, so a limit changed elsewhere shows as a change. Delete also sends the configured `reason`.
- **Match attributes** are strings only, as import fills them as strings.

## Resources

| Resource | Notes |
|---|---|
| `telnyx_authentication_provider` | SAML SSO settings. |
| `telnyx_auto_recharge_preferences` | Singleton; destroy keeps it. |
| `telnyx_user_profile` | Singleton (beta); destroy keeps it. Fields the API takes but does not return keep the configured value. |
| `telnyx_oauth_client` | Identified by `client_id`; scopes, grant types and redirect URIs are sets; `client_secret` is `create_only`. |
| `telnyx_managed_account` | No delete: destroy removes it from state and warns. `channel_limit` through `update_global_channel_limit`. API key and token are sensitive. |
| `telnyx_spend_limit` | See above. |
| `telnyx_enterprise` | Replaced with PUT. |
| `telnyx_display_identity_record` | Created under an enterprise, read by its own id; imports as `<enterprise_id>/<id>`. Call reasons are strings; a read hook flattens the objects the API returns. |
| `telnyx_display_identity_record_phone_number` | One number per resource. |
| `telnyx_enterprise_reputation` | Identified by the enterprise; frequency, webhook and LOA document through their own update operations. |
| `telnyx_enterprise_reputation_number` | One number per resource. |

The DIR workflow (submit, references, email verification, infringement update), branded calling activation, ownership confirmation, refresh and remediation are actions. Lists and read-only lookups are data sources; `telnyx_account_reputation_number` leaves out the `fresh` query parameter, because a fresh lookup is billed on every refresh.

## What is not covered

- `POST /dir/{dir_id}/loa`, `POST /dir/{dir_id}/bpo_loa` and `POST /enterprises/{enterprise_id}/reputation/loa` render a PDF to sign and upload. Not data Terraform can use: `out_of_scope`.
- API key spend limits, as above.
- `webhook_url` of reputation monitoring cannot be cleared by removing it: the webhook update has no clear operation, so the value is kept. Set another URL to change it.
- Terraform actions return no values: the remediation and external verification submissions are actions, and their status is read with the matching data sources.

## Testing

Codegen tests cover match, query, boolean unions, collection operations, response fields that name the object, and singletons with an id. Engine tests cover the match lookup and query parameters. Hook tests cover each hook. The conformance test runs every wave 6 resource; its fake API now answers batches of one phone number with one object. Acceptance tests are not run (no account).
