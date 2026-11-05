# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 6 is the account: who can sign in, how the account pays and caps spend, the accounts it manages, the OAuth clients apps use, and the enterprise identity behind Branded Calling and phone number reputation. The coverage review assigned 11 resources to it, with their data sources and actions. Phone numbers, SIM cards, hosted messaging numbers and storage buckets were also listed under wave 6 but are built by other waves.

A few of these objects have shapes the generator could not handle yet: spend limits are identified by product and period and take the period as a query parameter; DIR phone numbers and reputation numbers are added in batches and the DIR numbers are removed by a request body; the spend limit body is a union told apart by a boolean; once-per-account objects carry an id of their own; an OAuth client's secret is returned only when it is created.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - `read_from.match` names further attributes the collection element must match; they are required, force replacement and join the import id.
  - `query` names attributes that update and delete operations also take as query parameters.
  - Body unions whose variants are told apart by a boolean (`unlimited: false` or `true`) merge into one object.
  - An operation on the collection path, with the create path's parameters and no more, treats them all as parents (a delete that names the object in its body).
  - A response field the override makes required names the object, like a create field (hooks send it).
  - `once: true` keeps a value only the create response returns, such as a client secret.
  - A `singleton` whose response has an id of its own uses the fixed id in its place.
  - Attributes of an extra update with no `clear` operation are optional and computed.
  - Hooks: `Delete` replaces the generic delete; a not-found error from `AfterRead` means the object is gone.
  - A data-source-only lookup can leave out a query parameter with `attributes.<param>.skip`.
- **Wave 6 resources, data sources, actions and the ephemeral `telnyx_telephony_credential_token_material`** (the key material to mint telephony credential tokens locally, never stored in state), generated from overrides. The list is in `tasks.md`.
- **Hooks** for DIR call reasons (sent as strings, returned as objects), DIR and reputation phone numbers (one number per resource, sent as a list of one; DIR numbers deleted by body) and spend limits (the request always carries the limit; an entry without a limit is gone).
- **Coverage.** Three LOA renders that answer with a PDF move to `out_of_scope`.

## Capabilities

### Modified Capabilities
- `codegen`: objects identified by several fields, query parameters on update and delete, boolean-tagged unions, collection operations, response fields that name the object, values returned once, singletons with their own id, extra updates without a clear, delete hooks and gone objects, skipped query lookups.

### New Capabilities
- `account`: telephony credential token material, authentication providers, auto recharge, the user profile, OAuth clients, managed accounts, spend limits, enterprises, Display Identity Records and phone number reputation.

## Impact

- `internal/engine`: `Op.Query`, `ReadFrom.Match`, `Attr.Once`, `Hooks.Delete`; read hooks may report an object gone.
- `internal/codegen`: override keys `query`, `read_from.match`, `attributes.<f>.once`; union merging, collection operations, singleton ids, extra updates without a clear.
- `generator/overrides`: 44 new override files. `internal/provider/generated`: regenerated, no existing descriptor changes.
- `internal/provider`: four hook files and conformance test support for batches of phone numbers.
- `catalog/coverage.yaml`: three entries move to `out_of_scope`.
- Docs and examples for the new resources; a changelog entry.
