## ADDED Requirements

### Requirement: Forced deletes
An override's `force_destroy` SHALL name query parameters that take a boolean (alone or as a variant of a union) of the resource's delete operation, and SHALL add a local, optional boolean attribute `force_destroy`. When `force_destroy` is true in state, destroy SHALL send each named parameter as `true`; otherwise it SHALL send none. Generation SHALL fail when the override has no delete, or a name is not a boolean query parameter of the delete operation.

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

### Requirement: Lowercase values
An attribute override with `lowercase: true` SHALL reject a configured value that has capital letters or leading or trailing spaces; for a list or set of strings, each element. The error SHALL show at plan time.

#### Scenario: Mixed case address
- **WHEN** a configuration sets `to = "Someone@Example.com"` on an attribute with `lowercase: true`
- **THEN** plan fails with an error saying the value must be lowercase

### Requirement: Soft-deleted objects
Objects the API soft-deletes SHALL have a read hook that answers with a not-found error when the object's status says it is deleted, so read removes it from state (see "Delete hooks and gone objects").

#### Scenario: Removed suppression
- **WHEN** `GET /email_blocks/{id}` answers with `status = "removed"`
- **THEN** the suppression leaves state and the next plan creates it again

### Requirement: Local changes without an update
When a resource has no update operation and only local attributes (such as `force_destroy`) change, apply SHALL store the new values without a request.

#### Scenario: Set force_destroy on a membership
- **WHEN** `force_destroy = true` is added to an existing `telnyx_email_ip_pool_ip`
- **THEN** the plan updates it in place, apply sends no request, and destroy then sends `force=true&drain=true`
