# provider-core Specification

## Purpose
Configures the Telnyx provider and defines the behaviour every resource and data source shares when it talks to the Telnyx API.

## Requirements

### Requirement: Provider configuration
The provider SHALL be named `telnyx` and accept optional `api_key` (sensitive), `base_url` and `max_retries` attributes. Each SHALL fall back to `TELNYX_API_KEY`, `TELNYX_BASE_URL` and `TELNYX_MAX_RETRIES`. The default base URL SHALL be `https://api.telnyx.com/v2` and the default retry count 4.

#### Scenario: Key from the environment
- **WHEN** `api_key` is not set and `TELNYX_API_KEY` is set
- **THEN** the provider uses the environment value

#### Scenario: Missing key
- **WHEN** neither `api_key` nor `TELNYX_API_KEY` is set
- **THEN** configuration fails with an error naming both

#### Scenario: Unknown key at plan time
- **WHEN** `api_key` is set from a value only known after apply
- **THEN** configuration fails with an error explaining that the key must be known

### Requirement: Authentication
Every API request SHALL send the key as `Authorization: Bearer <key>` and a `User-Agent` of `terraform-provider-telnyx/<version>`. The key SHALL never appear in logs, plans or error messages.

#### Scenario: Request headers
- **WHEN** the provider sends any request
- **THEN** it carries the bearer token and the provider user agent

### Requirement: Retries
The provider SHALL retry HTTP 429 for every method, and transport errors and HTTP 500, 502, 503 and 504 only for GET, HEAD, PUT, DELETE and OPTIONS, up to the configured retry count. Waits SHALL use exponential backoff with jitter, capped at 30 seconds, and SHALL honour `Retry-After`.

#### Scenario: Rate limited create
- **WHEN** a POST receives 429 and then succeeds
- **THEN** the provider retries and the create succeeds

#### Scenario: Server error on create
- **WHEN** a POST receives 502
- **THEN** the provider does not retry and reports the error

#### Scenario: Retries exhausted
- **WHEN** a GET keeps receiving 503
- **THEN** the provider gives up after the configured retries and reports the last error

#### Scenario: Cancelled operation
- **WHEN** Terraform cancels the operation while the provider waits to retry
- **THEN** the provider stops waiting and returns the cancellation

### Requirement: Error reporting
API errors SHALL be reported with the HTTP method, path, status, each Telnyx error code and detail, and the request id when the API returns one.

#### Scenario: Not found
- **WHEN** the API returns 404 with a Telnyx error body
- **THEN** the error message contains the status, the Telnyx code and detail, and the request id
- **AND** resources can recognise it as not found and remove the object from state
