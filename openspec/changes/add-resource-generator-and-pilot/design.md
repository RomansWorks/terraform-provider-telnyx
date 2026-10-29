# Design

## Context

The spec has 2,555 schemas, with 154 `oneOf`, 202 `anyOf`, 32 discriminators and 316 free-form objects (`additionalProperties`). Request and response shapes differ (AI tools send `webhook: {...}` and get back `tool_definition`). Some responses are wrapped in `data`, others are not. Update verbs vary (POST, PUT, PATCH). Some objects cannot be deleted, some are created by upload, some complete asynchronously.

## Goals / Non-Goals

**Goals:** one tested engine shared by all resources; resource behaviour declared, not hand-coded; regeneration is deterministic; hand-written code only where the API genuinely needs it.

**Non-Goals:** rollout beyond the pilot; list data sources (plural); Terraform actions and ephemeral resources. Each comes in a later change.

## Decisions

### 1. Generated descriptors plus one shared engine, not generated CRUD per resource

The generator emits, per resource, a Go file holding a descriptor literal: operations, id and path parameters, envelope, and each attribute with its type, flags and JSON name. `internal/engine` builds the Plugin Framework schema from the descriptor and implements create, read, update, delete and import once, converting between Terraform values and JSON.

*Alternatives:* generating full CRUD code per resource (the Stainless and HashiCorp codegen approach) multiplies code across 168+ resources, so each fix has to be regenerated everywhere and the test surface grows. Generating framework schema code as well as the descriptor would duplicate the same facts in two forms that can drift. A schema loaded at runtime from a data file would be invisible to the compiler. A Go literal per resource is reviewable in diffs, type-checked, and the schema it yields is exactly what `tfplugindocs` reads from the provider.

### 2. Type mapping

| OpenAPI | Terraform |
|---|---|
| string, integer, number, boolean | String, Int64, Float64, Bool |
| enum | String with a one-of validator |
| `format: date-time` | String (RFC 3339) |
| object with properties | Single nested attribute |
| array of scalars | List (Set when an override says order does not matter) |
| array of objects | List nested attribute |
| `oneOf`/`anyOf` whose object variants each fix `type` to one value (the Telnyx convention, for example assistant tools) | One nested object with the union of the variants' fields; `type` selects the variant |
| `oneOf`/`anyOf` of one scalar type, or of one schema and `null` | That type |
| free-form object, `additionalProperties` without `properties`, other unions, arrays of arrays | String holding JSON, compared after normalisation so formatting and key order do not cause diffs |
| `['T', 'null']` | Optional T |

Nesting deeper than four levels is also rendered as normalised JSON, which keeps schemas usable for the deepest assistant settings (conversation flow nodes).

### 3. Required, optional and computed

- Required in the create request → Required.
- Optional in the request and returned by read → Optional and Computed, with `UseStateForUnknown`, so server defaults do not show as diffs.
- Returned by read but not in any request → Computed.
- Sent but never returned (for example an uploaded file) → kept from configuration. The engine cannot detect drift for these; docs say so.
- Path parameters of the item path become Required attributes that force replacement (for example `phone_number_id`).
- Accepted only by update (for example a number order's `regulatory_requirements`) → Optional; the engine sends them in an update straight after create.
- Accepted only by a PUT update and returned by read (a PUT body that repeats the whole object, server fields included) → Computed, and sent back unchanged.
- JSON attributes keep the configured value when the API returns the same document with extra keys (server defaults). A changed value in a key the user set is drift.

### 4. Envelope and ids

When a response schema's only property is `data`, the engine unwraps it. The id comes from the response field named by the descriptor (default `id`). Import takes the id, or `parent_id/id` when the item path has more than one parameter.

### 5. Update semantics

PATCH sends only changed attributes. POST updates are treated like PATCH. PUT sends the full object. Attributes the API cannot update (absent from every update request) force replacement.

### 6. Hooks for what the spec cannot express

A resource can register hooks in hand-written `internal/provider/hooks_<name>.go`: plan changes, before request, custom create (upload), after read, wait-until-ready and the destroy warning. Pilot hooks:

- `telnyx_ai_avatar`: create uploads `image_path` as multipart; the plan hook hashes the file into `image_sha256` and replaces the avatar when the content changes (a new path to the same content changes nothing).
- `telnyx_ai_tool`: after read, maps `tool_definition` back onto the variant attribute so drift on a webhook URL is detected; the plan hook rejects a variant that does not match `type` and replaces the tool when `type` changes.
- `telnyx_number_order`: after create, polls `GET /number_orders/{id}` until `status` is `success` (fails on `failure`), within a `timeouts.create` default of 10 minutes; destroy names the numbers that stay on the account.

### 7. Objects that cannot be deleted, and adopted settings

When a descriptor has no delete operation, destroy removes the object from state and adds a warning naming what still exists in Telnyx. Settings resources (`telnyx_phone_number_voice_settings`) adopt on create (PATCH), read with GET, and on destroy leave the settings as they are and only remove them from state.

### 8. Tests

- Generator: golden-file tests on a fixture spec, plus a CI check that `make generate` leaves no diff.
- Engine: unit tests for type conversion, envelope handling, PATCH diffs, 404-on-read removal and missing-delete warnings.
- Resources: unit tests through `terraform-plugin-testing` against an `httptest` fake of the Telnyx API that follows the spec's response shapes and adds server defaults (no credentials needed); acceptance tests per resource with an import step and an empty-plan check; sweepers for leaked test objects.

## Risks / Trade-offs

- [Kept-from-config attributes hide drift] → limited to fields the API never returns; listed in each resource's docs.
- [Normalised JSON attributes are less friendly than nested attributes] → used only where the spec gives no stable shape; an override can promote a field to a nested attribute when Telnyx documents its shape.
- [Importing a JSON attribute brings in server defaults] → after `terraform import`, the first plan may update a JSON attribute (such as an AI tool's `webhook`) to exactly the configured document. The update is harmless and the plan is empty afterwards.
- [Merged unions accept fields of other variants] → the API rejects mismatched fields; a per-resource plan hook can check them where it matters (AI tools do).
- [Generic engine bugs affect every resource] → the engine carries the heaviest unit tests, and each pilot pattern has a dedicated test.
- [The number order test buys a number] → gated by `TELNYX_ACC_BILLABLE=1`, run on demand only, and the sweeper releases the number.

## Open Questions

- Telnyx's spec does not say which assistant fields are server-defaulted; the first acceptance run will show which need to be Optional and Computed, and overrides will record them.
