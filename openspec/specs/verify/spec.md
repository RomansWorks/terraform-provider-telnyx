# verify Specification

## Purpose
Manages Telnyx Verify, which sends and checks verification codes: Verify profiles and their message templates.

## Requirements

### Requirement: Verify profiles
The provider SHALL manage `telnyx_verify_profile` and `telnyx_verify_profile_message_template`. A message template SHALL be read from the template list, and because the API cannot delete templates, destroy SHALL remove it from state with a warning.

#### Scenario: Template gone
- **WHEN** a message template is no longer in `GET /verify_profiles/templates`
- **THEN** the next plan proposes to create it again
