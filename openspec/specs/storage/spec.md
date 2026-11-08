# storage Specification

## Purpose
How the provider manages Telnyx Storage beyond buckets: KV namespaces and keys, SQL databases, CloudFS filesystems, migrations and bucket SSL certificates. Buckets themselves are created with the AWS provider against Telnyx's S3 endpoint.

## Requirements

### Requirement: KV keys
The provider SHALL manage `telnyx_storage_kv_key`. Its `value` SHALL be sensitive and sent as the raw body of `PUT /storage/kvs/{id}/keys/{key}`, with `ttl_secs` as a query parameter when set. Create and update SHALL both use that PUT, and read SHALL compare the raw value. A key SHALL be identified by `key` and imported as `<namespace_id>/<key>`.

#### Scenario: Write a value
- **WHEN** a key is created with `value = "hello"` and `ttl_secs = 3600`
- **THEN** the provider sends `PUT /storage/kvs/{id}/keys/{key}?ttl_secs=3600` with body `hello` and the next plan is empty

#### Scenario: Value changed outside Terraform
- **WHEN** the stored value differs from configuration
- **THEN** the next plan updates the key in place

### Requirement: KV namespaces
The provider SHALL manage `telnyx_storage_kv_namespace`. A namespace has no update, so `name` SHALL force replacement. `data "telnyx_storage_kv_keys"` SHALL read one page of a namespace's keys, narrowed by `prefix`.

#### Scenario: Rename a namespace
- **WHEN** a namespace's `name` changes
- **THEN** the plan replaces the namespace

### Requirement: SQL databases and CloudFS
The provider SHALL manage `telnyx_storage_sql_database` (no update; `name` forces replacement) with the `telnyx_storage_sql_database_query` action, and `telnyx_storage_cloudfs_filesystem`, whose `meta_token` SHALL be sensitive and kept from the create response, with the `telnyx_storage_cloudfs_rotate_meta_token` action.

#### Scenario: Database lifecycle
- **WHEN** a database is created, destroyed and imported
- **THEN** each step uses the matching operation and import yields an empty plan

### Requirement: Migrations
The provider SHALL manage `telnyx_storage_migration_source`, whose `provider` SHALL be `source_provider` and `provider_auth` sensitive, and `telnyx_storage_migration`. Neither has an update: every argument SHALL force replacement. The API cannot delete migrations: destroy SHALL remove a migration from state and warn; the `telnyx_storage_migration_stop` action stops one.

#### Scenario: Destroy a migration
- **WHEN** a migration is destroyed
- **THEN** no request is sent, it leaves state, and a warning says it still exists in Telnyx

### Requirement: Bucket SSL certificates
The provider SHALL manage `telnyx_storage_bucket_ssl_certificate`, identified by `bucket_name`. `certificate_pem` and `private_key_pem` (sensitive) SHALL be uploaded as the `certificate` and `private_key` files of a multipart form, on create and on update. Buckets themselves SHALL NOT be resources of this provider: the docs SHALL show how to create them with the AWS provider pointed at Telnyx's S3 endpoint.

#### Scenario: Upload
- **WHEN** a certificate is created for bucket `assets`
- **THEN** `PUT /storage/buckets/assets/ssl_certificate` carries both PEM files and the next plan is empty

### Requirement: Storage provisioning
Create of `telnyx_storage_kv_namespace` and `telnyx_storage_sql_database` SHALL wait, within the `create` timeout, until `status` is `provision_ok`, because their keys and queries answer 409 before then. `provision_failed`, `deleting` and `delete_failed` SHALL fail the apply, leaving the object in state.

#### Scenario: Namespace and key in one apply
- **WHEN** a namespace answers `pending` and then `provision_ok`
- **THEN** create returns after `provision_ok`, and a key in the same configuration is written after it
