# Spec Delta

## Purpose

Manages Telnyx voice configuration such as outbound voice profiles.

## ADDED Requirements

### Requirement: Outbound voice profile
The provider SHALL manage `telnyx_outbound_voice_profile` with `POST /outbound_voice_profiles`, `GET`, `PATCH` and `DELETE /outbound_voice_profiles/{id}`, reading objects from the `data` envelope. `tags` and `whitelisted_destinations` SHALL be sets.

#### Scenario: Lifecycle
- **WHEN** a profile is created, its `concurrent_call_limit` changed, then destroyed
- **THEN** create, PATCH and DELETE are called and each follow-up plan is empty

#### Scenario: Import
- **WHEN** a profile is imported by id
- **THEN** a matching configuration has an empty plan
