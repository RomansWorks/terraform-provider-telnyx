## ADDED Requirements

### Requirement: Forced deletes
An override's `force_destroy` SHALL name boolean query parameters of the resource's delete operation, and SHALL add a local, optional boolean attribute `force_destroy`. When `force_destroy` is true in state, destroy SHALL send each named parameter as `true`; otherwise it SHALL send none. Generation SHALL fail when the override has no delete, or a name is not a boolean query parameter of the delete operation.

#### Scenario: Force a delete
- **WHEN** a resource with `force_destroy: [force, drain]` and `force_destroy = true` in state is destroyed
- **THEN** the delete request carries `drain=true&force=true`

#### Scenario: Plain delete
- **WHEN** the same resource without `force_destroy` is destroyed
- **THEN** the delete request carries no query parameters

#### Scenario: Unknown parameter
- **WHEN** an override names `force_destroy: [purge]` and the delete operation has no boolean query parameter `purge`
- **THEN** `make generate` fails, naming the parameter

### Requirement: Union list responses
When a list operation's response has no `data` array but is a `oneOf` or `anyOf` of objects, the generator SHALL take the items from the first variant whose `data` is an array. Paging SHALL be derived from the operation's parameters as for other lists.

#### Scenario: Two page shapes
- **WHEN** `GET /email_blocks` answers with `oneOf` a page-number shape and a cursor shape
- **THEN** `data "telnyx_email_suppressions"` lists the items, paged by `page[number]`
