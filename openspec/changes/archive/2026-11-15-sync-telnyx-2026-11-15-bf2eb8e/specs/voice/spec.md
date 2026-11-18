## ADDED Requirements

### Requirement: TeXML secrets
The provider SHALL manage `telnyx_texml_secret`, identified by `name`, with the secret value in the write-only `value_wo`. Create SHALL send `POST /texml/secrets`. Because the API can neither read nor delete a secret, changing `name` or `value_wo_version` SHALL replace it, refresh SHALL keep state, import SHALL be refused, and destroy SHALL remove it from state and warn that it stays in the account.

#### Scenario: Rotate the value
- **WHEN** `value_wo` and `value_wo_version` change
- **THEN** the provider sends `POST /texml/secrets` with the same `name` and the new value, and no delete

#### Scenario: Destroy keeps the secret
- **WHEN** a TeXML secret is destroyed
- **THEN** no request is sent, it leaves state, and a warning says it still exists in the account
