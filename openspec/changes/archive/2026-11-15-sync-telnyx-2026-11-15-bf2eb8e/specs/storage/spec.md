## ADDED Requirements

### Requirement: Presigned object URLs
The provider SHALL offer the ephemeral `telnyx_storage_object_presigned_url` for one object, identified by `bucket_name` and `object_name`, created with `POST /storage/buckets/{bucketName}/{objectName}/presigned_url` and an optional `ttl` in seconds. The URL, its token and expiry SHALL be under `content` and SHALL never be stored in plan or state.

#### Scenario: One-hour link
- **WHEN** the ephemeral is opened with `ttl = 3600`
- **THEN** the request body is `{"ttl": 3600}` and `content.presigned_url` holds the link
