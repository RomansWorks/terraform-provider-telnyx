# Proposal

## Why

The coverage review set aside three kinds of operations that are not resources, and one kind of resource that needs a safer destroy:

- **121 actions.** These are one-shot operations such as cloning an assistant, appealing a campaign or registering SIM cards. They change something in Telnyx but leave no object to manage.
- **3 ephemeral reads.** These return a secret: a SIM card activation code, a telephony credential token and a WireGuard peer config. The value must never enter state.
- **Secret request fields.** Some create and update bodies take a password, an API key or a private key. Today these are stored in state, marked sensitive.
- **Numbers and SIM cards.** Roman decided that destroy keeps them in Telnyx by default, and that `release_on_destroy = true` opts in to the real delete, which releases or deactivates them.

Waves 1 to 6 need all four, so they are built now. This is part 2 of step 2. Part 1 (`add-collection-reads`) covered lists, reading from a collection and extra update endpoints.

This change covers no upstream range: it builds on the current pin, `team-telnyx/openapi@3d1ee2b`.

## What Changes

- **Actions.** An override with `kind: action` and one `op` produces a Terraform action (Terraform 1.14 and later). Its path parameters and request body are its configuration. Invoking it sends the request once.
- **Ephemeral resources.** An override with `kind: ephemeral` and an `open` operation produces an ephemeral resource. Its path parameters are required and its response fields are computed and sensitive. Nothing is written to state or plan.
- **Write-only secrets.** An attribute override with `write_only: true` makes the field `<name>_wo`. Terraform never stores it, and a companion `<name>_wo_version` controls when it is sent again.
- **Release on destroy.** An override with `release_on_destroy: true` adds an optional `release_on_destroy` attribute, false by default. Destroy calls the delete operation only when it is true. Otherwise the object leaves state and a warning says it is still in Telnyx.

No existing resource or data source changes behaviour. The pilots get no new attributes, so nothing here is user-visible until waves 1 to 6 use it, and there is no changelog entry.

## Capabilities

### Modified Capabilities
- `codegen`: actions, ephemeral resources, write-only secrets and release on destroy.

## Impact

- `internal/engine`: action and ephemeral descriptors and their framework types, write-only attributes read from configuration, and a conditional delete.
- `internal/codegen`: the override keys `kind: action`, `kind: ephemeral`, `op`, `open`, `write_only` and `release_on_destroy`, plus fixtures and goldens.
- `internal/provider`: registers generated actions and ephemeral resources.
