# media Specification

## Purpose
TBD - created by archiving change add-wave-5-ai-storage-media. Update Purpose after archive.

## Requirements

### Requirement: Media files
The provider SHALL manage `telnyx_media`, identified and imported by `media_name`, created from `media_url` with the JSON form of `POST /media`. Its list data source SHALL be `telnyx_media_files`.

#### Scenario: Replace the file
- **WHEN** `media_url` changes
- **THEN** the provider sends `PUT /media/{media_name}` and the next plan is empty

### Requirement: Video rooms
The provider SHALL manage `telnyx_room`. A room's sessions SHALL not be attributes: they are its call history.

#### Scenario: Room lifecycle
- **WHEN** a room is created, its `max_participants` changed with PATCH, destroyed and imported
- **THEN** each step uses the matching operation and every plan after apply is empty
