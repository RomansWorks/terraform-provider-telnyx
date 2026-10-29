# Spec Delta

## Purpose

Defines how Terraform resources and data sources are derived from the pinned Telnyx OpenAPI spec and per-resource overrides, and the behaviour all generated resources share.

## ADDED Requirements

### Requirement: Resources are generated from the spec and overrides
Each resource SHALL be defined by an override file in `generator/overrides/` naming its Terraform type and its operations as `METHOD /path`. Its attributes SHALL come from the operations' request and response schemas in the pinned spec, adjusted only by the override.

#### Scenario: Regeneration is deterministic
- **WHEN** `make generate` runs twice with no change to the spec or overrides
- **THEN** the second run changes no file

#### Scenario: Upstream adds a field
- **WHEN** a sync adds an optional field to a resource's request and response schemas
- **THEN** `make generate` adds an optional attribute for it with no override change

### Requirement: Attribute derivation
Attributes SHALL be Required when required by the create request, Optional and Computed when optional in the request and returned by read, Computed when only returned, and kept from configuration when sent but never returned. Attributes missing from every update request SHALL force replacement.

#### Scenario: Server default
- **WHEN** a user omits an optional attribute and the API returns a default value
- **THEN** the plan after apply is empty

#### Scenario: Update-only field
- **WHEN** a user sets an attribute that only the update operation accepts
- **THEN** the provider sends it in an update straight after create

#### Scenario: Immutable field
- **WHEN** a user changes an attribute that no update operation accepts
- **THEN** the plan replaces the resource

### Requirement: Polymorphic and free-form values
Unions whose object variants each fix `type` to one value SHALL become one nested object with the union of the variants' fields, where `type` selects the variant. Free-form objects and other unions SHALL become strings holding JSON, compared after normalisation. A JSON attribute SHALL keep its configured value when the API returns the same document with extra keys.

#### Scenario: Tagged union
- **WHEN** an assistant configures `tools = [{ type = "hangup", hangup = {} }]`
- **THEN** the request carries `{"type": "hangup", "hangup": {}}`

#### Scenario: JSON formatting
- **WHEN** a JSON attribute differs from the API value only in key order or whitespace
- **THEN** the plan shows no change

#### Scenario: Server defaults in JSON
- **WHEN** the API returns a JSON attribute with keys the user did not set, and the keys the user set are unchanged
- **THEN** the plan shows no change

### Requirement: Response envelopes
When a response schema's only property is `data`, the provider SHALL read the object from `data`.

#### Scenario: Wrapped response
- **WHEN** the API returns `{"data": {...}}`
- **THEN** the resource's attributes come from the inner object

### Requirement: Shared lifecycle behaviour
Every generated resource SHALL remove itself from state when read returns 404, support import by id (or `parent_id/id` for nested paths), and send only changed attributes on PATCH or POST updates and the full object on PUT updates.

#### Scenario: Deleted outside Terraform
- **WHEN** an object was deleted in the Telnyx portal
- **THEN** the next plan proposes to create it again

#### Scenario: Import
- **WHEN** a user runs `terraform import telnyx_ai_assistant.x <id>`
- **THEN** the resource is read into state and the next plan is empty for a matching configuration

### Requirement: Objects without delete
When the API offers no delete for an object, destroy SHALL remove it from state and warn that the object still exists in Telnyx.

#### Scenario: Destroy without delete
- **WHEN** a resource with no delete operation is destroyed
- **THEN** no API call is made, the object leaves state, and a warning names it

### Requirement: Data sources
A resource's read operation SHALL also be exposed as a data source of the same name that looks the object up by id, with every attribute Computed.

#### Scenario: Look up by id
- **WHEN** a configuration declares `data "telnyx_ai_assistant" "x" { id = "..." }`
- **THEN** its attributes are populated from `GET /ai/assistants/{assistant_id}`
