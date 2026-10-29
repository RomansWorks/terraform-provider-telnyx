# Spec Delta

## Purpose

Manages custom avatars used by Telnyx AI assistants.

## ADDED Requirements

### Requirement: AI avatar upload
The provider SHALL manage `telnyx_ai_avatar` by uploading the local file at `image_path` with `POST /ai/avatars` as multipart form data. A computed `image_sha256` of the uploaded file SHALL force replacement when the file's content changes; a new path to the same content SHALL change nothing in Telnyx. `name` SHALL update in place with `PATCH /ai/avatars/{avatar_id}`.

#### Scenario: Rename
- **WHEN** `name` changes
- **THEN** the avatar is updated in place

#### Scenario: New image
- **WHEN** the file at `image_path` changes
- **THEN** the plan replaces the avatar, and a warning says the previous avatar still exists in Telnyx

#### Scenario: Import
- **WHEN** an avatar is imported by id
- **THEN** it is read into state with `image_path` unset, and a plan with `image_path` configured records the file's hash without replacing the avatar

### Requirement: AI avatar destroy
Because the API has no delete for avatars, destroy SHALL remove the avatar from state and warn that it remains in Telnyx.

#### Scenario: Destroy
- **WHEN** an avatar is destroyed
- **THEN** it leaves state with a warning and no API call is made
