## ADDED Requirements

### Requirement: Sign-in and API access
The provider SHALL manage `telnyx_authentication_provider` (SAML single sign-on) and `telnyx_oauth_client`. An OAuth client SHALL be identified by `client_id`; its scopes, grant types and redirect URIs SHALL be sets, and its `client_secret` SHALL be sensitive and kept from the create response.

#### Scenario: Confidential client
- **WHEN** a `telnyx_oauth_client` with `client_type = "confidential"` is created
- **THEN** `client_secret` is set in state and marked sensitive, and the next plan is empty

### Requirement: Account settings
The provider SHALL manage `telnyx_auto_recharge_preferences` and `telnyx_user_profile` as objects that exist once per account (per user for the profile). Destroy SHALL remove them from state and warn.

#### Scenario: Auto recharge
- **WHEN** `telnyx_auto_recharge_preferences` sets `enabled = true` and a threshold and recharge amount
- **THEN** `PATCH /payment/auto_recharge_prefs` is sent and the next plan is empty

### Requirement: Spend limits
The provider SHALL manage `telnyx_spend_limit` per product and period, with exactly one of `amount` and `unlimited = true` sent on every create and update. Destroy SHALL remove the limit with `DELETE /spend_limits/{product}?period=<period>`. An entry without a limit SHALL read as gone. Limits of single API keys are not managed.

#### Scenario: Change the reason only
- **WHEN** only `reason` of a limit with `amount = 50` changes
- **THEN** `PATCH /spend_limits/{product}?period=<period>` carries `amount` 50 and the reason

#### Scenario: Limit removed outside Terraform
- **WHEN** the list returns the product and period with `limit: null`
- **THEN** the next plan creates the limit again

### Requirement: Managed accounts
The provider SHALL manage `telnyx_managed_account`, with its global channel limit set through `PATCH /managed_accounts/{id}/update_global_channel_limit`, and its API key, API token and password sensitive. Because the API cannot delete managed accounts, destroy SHALL remove it from state and warn. `telnyx_managed_account_enable` and `telnyx_managed_account_disable` SHALL be actions.

#### Scenario: Destroy
- **WHEN** a managed account is destroyed
- **THEN** no request is sent and a warning says it still exists in Telnyx

### Requirement: Enterprises and Branded Calling
The provider SHALL manage `telnyx_enterprise`, `telnyx_display_identity_record` (created under an enterprise, imported as `<enterprise_id>/<id>`) and `telnyx_display_identity_record_phone_number` (one number per resource, added and removed through the DIR's batch operations). A DIR's call reasons SHALL be strings. The DIR workflow (submit, references, email verification, infringement update) and branded calling activation SHALL be actions.

#### Scenario: Call reasons
- **WHEN** a DIR is created with `call_reasons = ["Appointment reminders"]` and the API returns `[{"reason": "Appointment reminders", "created_at": "..."}]`
- **THEN** the next plan is empty

#### Scenario: Remove a number
- **WHEN** a `telnyx_display_identity_record_phone_number` for `+13125550100` is destroyed
- **THEN** `DELETE /dir/{dir_id}/phone_numbers` is sent with `{"phone_numbers": ["+13125550100"]}`

### Requirement: Phone number reputation
The provider SHALL manage `telnyx_enterprise_reputation`, identified by its enterprise, with `check_frequency` and `webhook_url` set through their own update operations, and `telnyx_enterprise_reputation_number`, identified by its phone number. Refresh, remediation, ownership confirmation, external verification and LOA replacement SHALL be actions; their results are read with data sources.

#### Scenario: Change the frequency
- **WHEN** only `check_frequency` changes
- **THEN** only `PATCH /enterprises/{enterprise_id}/reputation/frequency` is sent
