# release Specification

## Purpose
Defines how provider versions are numbered, how breaking changes are introduced, and how releases are built and published.

## Requirements

### Requirement: Semantic versioning
Releases SHALL follow semantic versioning. Adding resources, data sources or optional attributes is a minor release; fixes are patch releases; removing or renaming a resource or attribute, changing a type, or changing an import id requires a major release.

#### Scenario: New resource
- **WHEN** a release only adds resources
- **THEN** its version increases the minor number

### Requirement: Deprecation before removal
Anything to be removed SHALL first ship with a deprecation message naming its replacement, for at least one minor release, and be removed only in the next major release. Renamed resources SHALL support Terraform `moved` blocks, and schema changes SHALL include state upgraders so users do not recreate infrastructure.

#### Scenario: Upstream removes a field
- **WHEN** the Telnyx API removes a field that backs an attribute
- **THEN** the attribute is marked deprecated in the next minor release and removed in the next major release

### Requirement: Changelog
Every user-visible change SHALL add a changelog entry in the same pull request, and each release SHALL publish the collected entries in `CHANGELOG.md`.

#### Scenario: Pull request without entry
- **WHEN** a pull request changes provider behaviour without a changelog entry
- **THEN** review asks for one before merge

### Requirement: Signed releases
Pushing a `v*` tag SHALL build binaries for all supported platforms, publish SHA256 checksums signed with the project GPG key and the registry manifest, and make the release available to the Terraform Registry and the OpenTofu registry under `romansworks/telnyx`.

#### Scenario: Tag pushed
- **WHEN** `v1.2.0` is tagged
- **THEN** a GitHub release with signed checksums is created and the registries pick it up
