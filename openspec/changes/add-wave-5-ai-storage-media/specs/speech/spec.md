## ADDED Requirements

### Requirement: Voice clones
The provider SHALL manage `telnyx_voice_clone`, made from a voice design (`voice_design_id`, sent with `POST /voice_clones`) or from a recording (`audio_path`, uploaded with `POST /voice_clones/from_upload` as a multipart form with the other create arguments). Exactly one of the two SHALL be set. The recording's SHA-256 SHALL be tracked in `audio_sha256`: different content SHALL replace the clone, and a moved file with the same content SHALL not. A clone SHALL be read from the clone list.

#### Scenario: From a design
- **WHEN** a clone sets `voice_design_id`
- **THEN** the provider sends JSON to `POST /voice_clones` and the next plan is empty

#### Scenario: From a recording
- **WHEN** a clone sets `audio_path` to a WAV file
- **THEN** the provider uploads it with the clone's name, language and gender, and a new recording at the same path replaces the clone

#### Scenario: Rename
- **WHEN** `name` changes
- **THEN** the provider sends `PATCH /voice_clones/{id}`

### Requirement: Voice clone arguments
`provider` on `telnyx_voice_clone` SHALL be `voice_provider`, and the create request shapes of the different providers SHALL be merged into one set of arguments.

#### Scenario: Provider argument
- **WHEN** a clone sets `voice_provider = "telnyx"`
- **THEN** the create request carries `provider: "telnyx"`

### Requirement: Voice designs and pronunciation dictionaries
The provider SHALL manage `telnyx_voice_design` (`provider` as `voice_provider`; a design's text and prompt force replacement, and only its name is updated) with the `telnyx_voice_design_version_delete` action, and `telnyx_pronunciation_dictionary`, whose items are alias or phoneme entries told apart by `type`.

#### Scenario: Rename a design
- **WHEN** a design's `name` changes
- **THEN** the provider sends `PATCH /voice_designs/{id}` and the next plan is empty
