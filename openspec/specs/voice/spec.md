# voice Specification

## Purpose
Manages Telnyx voice configuration such as outbound voice profiles.

## Requirements

### Requirement: Outbound voice profile
The provider SHALL manage `telnyx_outbound_voice_profile` with `POST /outbound_voice_profiles`, `GET`, `PATCH` and `DELETE /outbound_voice_profiles/{id}`, reading objects from the `data` envelope. `tags` and `whitelisted_destinations` SHALL be sets.

#### Scenario: Lifecycle
- **WHEN** a profile is created, its `concurrent_call_limit` changed, then destroyed
- **THEN** create, PATCH and DELETE are called and each follow-up plan is empty

#### Scenario: Import
- **WHEN** a profile is imported by id
- **THEN** a matching configuration has an empty plan

### Requirement: Connections
The provider SHALL manage `telnyx_credential_connection`, `telnyx_fqdn_connection`, `telnyx_ip_connection`, `telnyx_uac_connection`, `telnyx_external_connection` and `telnyx_mobile_voice_connection`. Passwords SHALL be masked (sensitive, kept in state when the API answers with asterisks), authentication tokens SHALL be sensitive, and `tags` SHALL be sets. Codec lists SHALL stay lists, because their order is a preference.

#### Scenario: Sensitive password
- **WHEN** a credential connection is planned
- **THEN** its `password` is shown as sensitive

### Requirement: Settings of connections
The provider SHALL manage `telnyx_fqdn_connection_authentication`, `telnyx_external_connection_phone_number_settings`, `telnyx_custom_storage_credential` and `telnyx_dialogflow_connection` as settings of an existing connection or number, identified by its path parameters. Create SHALL apply the configured settings. Destroy SHALL delete them where the API can, and otherwise remove them from state with a warning.

#### Scenario: Import phone number settings of an external connection
- **WHEN** `terraform import telnyx_external_connection_phone_number_settings.x ec1/pn1` runs
- **THEN** the settings are read and a matching configuration has an empty plan

### Requirement: Applications and call features
The provider SHALL manage `telnyx_call_control_application`, `telnyx_texml_application`, `telnyx_fax_application`, `telnyx_queue`, `telnyx_texml_queue`, `telnyx_pay_connector`, `telnyx_siprec_connector`, `telnyx_telephony_credential` and `telnyx_mobile_push_credential`. A SIPREC connector's and a Pay connector's id SHALL be their `name`, a queue's its `queue_name`, and a TeXML queue's its `sid` under `account_sid`.

#### Scenario: Import a TeXML queue
- **WHEN** `terraform import telnyx_texml_queue.x AC1/QU1` runs
- **THEN** the queue is read from `GET /texml/Accounts/AC1/Queues/QU1`

### Requirement: Voice channels
The provider SHALL manage `telnyx_inbound_voice_channel_settings`, the account's US-zone channel count, as once-per-account settings, and `telnyx_voice_channel_zone` for a non-US zone identified by `channel_zone_id`, read from the zone list. Neither SHALL be deleted on destroy.

#### Scenario: Change zone channels
- **WHEN** `channels` of a voice channel zone changes from 5 to 10
- **THEN** the provider sends `PUT /channel_zones/{channel_zone_id}` with 10 and the next plan is empty

### Requirement: Telephony credential token
The provider SHALL offer the ephemeral `telnyx_telephony_credential_token`, whose `token` is a JWT for a WebRTC client minted with `POST /telephony_credentials/{id}/token`. It SHALL never be stored in plan or state.

#### Scenario: Pass a token to another provider
- **WHEN** a configuration passes `ephemeral.telnyx_telephony_credential_token.t.token` to a write-only argument
- **THEN** the token reaches it and is not in state

### Requirement: Voice actions and lookups
The provider SHALL offer the wave 1 voice actions (registration status checks, external connection uploads and location updates, Operator Connect refresh) and data sources (SIP registration status, external connection civic addresses, releases and uploads, channel billing numbers).

#### Scenario: Check registration
- **WHEN** `action "telnyx_uac_connection_check_registration_status"` is invoked for connection `u1`
- **THEN** the provider sends `POST /uac_connections/u1/actions/check_registration_status` once

### Requirement: Voice secrets
Storage credentials, credential secrets and push keys SHALL be sensitive. The Dialogflow service account (`service_account_wo`) and a Pay connector's `config` (`config_wo`) SHALL be write-only, because the API never returns them in full.

#### Scenario: Rotate a Pay connector's credentials
- **WHEN** `config_wo` changes together with `config_wo_version`
- **THEN** the update sends the new `config`, and state holds no `config_wo` value
