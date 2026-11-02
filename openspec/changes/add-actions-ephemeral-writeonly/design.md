# Design

As in part 1, everything is generated from overrides. The new keys say what the spec cannot: that an operation is an action or an ephemeral read, that a field is a secret, and that a destroy should keep the object.

## Actions

```yaml
type: telnyx_ai_assistant_clone
kind: action
description: Clones an AI assistant.
op: POST /ai/assistants/{assistant_id}/clone
```

- **Configuration.** Path parameters are required string attributes. They can be renamed with `path_params`, as in resources. The request body's fields follow, Required when the spec requires them and Optional otherwise. Nested objects and JSON fields are derived the same way as resource attributes, with the same `attributes` overrides.
- **Invoke.** The engine sends the operation once, with the fields that are set. It reports the HTTP outcome as a progress message. An API error fails the action with the same error text that resources show.
- **No results.** Terraform actions have no output values, so the response is not exposed. When a user needs what an operation returns (an id from clone, purchase or import), the operation belongs in a resource. Each wave decides that case by case and records it in `catalog/coverage.yaml`.
- **Versions.** Actions need Terraform 1.14. Older Terraform versions and OpenTofu do not read action schemas, so the provider still works there without them. The acceptance run checks this with OpenTofu.

## Ephemeral resources

```yaml
type: telnyx_sim_card_activation_code
kind: ephemeral
description: The activation code of a SIM card.
open: GET /sim_cards/{id}/activation_code
```

- Path parameters are required. Response fields are computed and sensitive.
- `Open` sends the request. There is no renew and no close, because these values do not expire on Telnyx's side.
- Ephemeral resources need Terraform 1.10 or OpenTofu 1.11.

## Write-only secrets

```yaml
attributes:
  password:
    write_only: true
```

- The attribute becomes `password_wo`: Optional (or Required when the spec requires it), WriteOnly and Sensitive. It is never Computed and never stored.
- The generator adds `password_wo_version`, an optional number. The engine never sends it. Terraform cannot see a change to a value it never stores, so the user bumps this number to send the secret again.
- **Create** sends the write-only value from configuration. **Update** sends it only when `_wo_version` changed. When no update operation accepts the field, a change to `_wo_version` replaces the object.
- Write-only attributes need Terraform 1.11 or OpenTofu 1.11. The README already asks for Terraform 1.13.
- There is no stored twin attribute: the provider has no users yet, so secrets are write-only from their first release.

## Release on destroy

```yaml
release_on_destroy: true
```

- This needs a delete operation. The generator adds `release_on_destroy`, an optional boolean, false by default. It is never sent to the API, and changing it never replaces the object.
- **Destroy with false:** no API call. The object leaves state, and a warning says that it is still in Telnyx and still billed, and how to release it.
- **Destroy with true:** the delete operation runs, as for any resource.
- Destroy reads the value from state, so the user sets it to true and applies before destroying. The generated description says so.

## Alternatives considered

- **Actions as resources** whose create sends the operation (a "trigger" resource). Rejected: such resources do nothing on later applies and confuse plans. Terraform added actions for exactly this case.
- **Keep secrets stored and sensitive.** Rejected for new secrets: sensitive values still sit in plain text in state files.
