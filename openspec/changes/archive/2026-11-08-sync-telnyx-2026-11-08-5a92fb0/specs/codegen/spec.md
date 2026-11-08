## REMOVED Requirements

### Requirement: Versioned spec paths
**Reason**: Its scenario named mobile phone numbers, whose spec paths no longer repeat `/v2`. Replaced by "Spec paths repeating the version", with the same behaviour.
**Migration**: None; requests are unchanged.

## ADDED Requirements

### Requirement: Spec paths repeating the version
A spec path that starts with `/v2/` SHALL be sent once under the base URL's `/v2`, not as `/v2/v2/...`.

#### Scenario: Path repeating the version
- **WHEN** a descriptor uses a spec path such as `GET /v2/example/{id}` with the default base URL
- **THEN** the request goes to `https://api.telnyx.com/v2/example/{id}`
