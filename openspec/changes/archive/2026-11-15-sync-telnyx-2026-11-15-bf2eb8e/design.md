# Design

## Unreadable resources (`read: none`)

`POST /texml/secrets` is the only operation for TeXML secrets: it creates a secret, or replaces the value when the name exists, and answers with the name and `"value": "REDACTED"`. There is no read, list or delete.

An override may set `read: none`. The generator then takes attributes from the create request and the create response, sets `Unreadable` on the descriptor, registers no data source, and rejects `update`, `updates`, `read_from` and `list` alongside it. The engine:

- creates with the create operation and builds state from its response, as any resource without a separate read;
- on refresh returns the state unchanged and sends nothing;
- refuses import with "Import not supported", since there is nothing to read;
- on destroy, having no delete, removes the object from state with the usual warning.

Without an update operation every change forces replacement, including `value_wo_version` (the existing rule for write-only versions). Replacement destroys first, which sends nothing, then creates, which replaces the value under the same name. That is why the secret needs no update operation even though the API would accept one.

Alternatives: an action would work only in Terraform 1.14 and not in OpenTofu, and returns nothing to reference. A hand-written resource would duplicate the engine. A generic flag is small and covers other create-only APIs.

## Ephemeral arguments

Ephemeral resources took only path parameters. Client tokens and presigned URLs take optional lifetimes in the request body. `DeriveEphemeral` now derives the request body fields as Optional (or Required) arguments, neither Computed nor Sensitive, with `InCreate`. A body field that is an object, or a list of objects, is a generator error, to keep the schema to values. A name used by both request and response is an error. `Open` sends a JSON body of the arguments the configuration sets (`{}` when none) whenever the operation has a body; operations without one are unchanged.

## Resource notes

- `telnyx_storage_object_presigned_url`: the response has no `data` envelope; its fields are under `content` (`presigned_url`, `token`, `expires_at`). The object name is a path parameter, so a `/` in it is sent percent-encoded.
- `telnyx_room_client_token`: `token`, `refresh_token` and their expiry times are sensitive results; refreshing is left to the client.

## Tests

- Codegen: `read: none` derivation and its validation errors; ephemeral request arguments and the object-field error.
- Engine: an ephemeral sends exactly the set arguments; an unreadable resource creates with one request, refreshes and destroys with none, replaces on a version change and refuses import.
- Conformance runs `telnyx_texml_secret` with an import step that expects the refusal.
