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
Pushing a `v*` tag SHALL build binaries for all supported platforms, publish SHA256 checksums signed with the project GPG key and the registry manifest, and make the release available to the Terraform Registry and the OpenTofu registry under `romansworks/telnyx`. Before building, the release workflow SHALL pass the release gate (`make api-check-release`), the unit tests and the generated-docs check, and SHALL fail without publishing anything if any of them fails.

#### Scenario: Tag pushed
- **WHEN** `v1.2.0` is tagged
- **THEN** a GitHub release with signed checksums is created and the registries pick it up

#### Scenario: Release gate fails
- **WHEN** a `v*` tag is pushed while an operation in `catalog/coverage.yaml` is unreviewed
- **THEN** the release workflow fails before GoReleaser runs and no GitHub release is created

### Requirement: Supported CLIs
The provider SHALL work with Terraform 1.11 or later and OpenTofu 1.11 or later. Actions SHALL need Terraform 1.14 or later and are not available in OpenTofu, which has no `action` blocks; everything else, including ephemeral resources and write-only attributes, SHALL work in both. CI SHALL run the unit tests and validate the examples with both CLIs, without Telnyx credentials, and the documentation SHALL say which features each CLI supports.

#### Scenario: OpenTofu user
- **WHEN** a configuration uses resources, data sources, ephemeral resources or write-only attributes with OpenTofu 1.11
- **THEN** it validates, plans and applies as with Terraform

#### Scenario: Action in OpenTofu
- **WHEN** an OpenTofu configuration declares an `action` block
- **THEN** OpenTofu rejects the block, and the guide explains that actions need Terraform 1.14

#### Scenario: Broken example
- **WHEN** a pull request adds an example that does not validate with either CLI
- **THEN** CI fails

### Requirement: Release notes
Each release SHALL have a collated changelog section, `.changes/<version>.md` (produced by `changie batch <version>`), merged into `CHANGELOG.md` before its tag is pushed. The GitHub release SHALL use that section as its release notes.

#### Scenario: Tag without collated changelog
- **WHEN** `v0.2.0` is tagged but `.changes/0.2.0.md` does not exist
- **THEN** the release workflow fails before GoReleaser runs

#### Scenario: Release notes published
- **WHEN** `v0.1.0` is released
- **THEN** the GitHub release body is the contents of `.changes/0.1.0.md`

### Requirement: Local release check
`make release-snapshot` SHALL build the release artifacts for every platform without signing or publishing, so maintainers can verify the GoReleaser configuration before tagging.

#### Scenario: Snapshot build
- **WHEN** a maintainer runs `make release-snapshot`
- **THEN** zipped binaries, checksums and the registry manifest are written to `dist/` and nothing is signed or uploaded
