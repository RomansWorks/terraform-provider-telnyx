## ADDED Requirements

### Requirement: Idempotency keys
Every POST, PUT, PATCH and DELETE request SHALL carry an `Idempotency-Key` header holding a random UUID. Retries of one request SHALL send the same key; separate requests SHALL send different keys. GET requests SHALL carry none. A hook MAY set request headers for one request, overriding the client's.

#### Scenario: Retried create
- **WHEN** `POST /storage/cloudfs` is answered with 429 and retried
- **THEN** both attempts carry the same `Idempotency-Key`

#### Scenario: Separate changes
- **WHEN** two token keys are created
- **THEN** the two requests carry different keys
