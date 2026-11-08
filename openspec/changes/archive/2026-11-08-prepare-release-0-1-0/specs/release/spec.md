## MODIFIED Requirements

### Requirement: Signed releases
Pushing a `v*` tag SHALL build binaries for all supported platforms, publish SHA256 checksums signed with the project GPG key and the registry manifest, and make the release available to the Terraform Registry and the OpenTofu registry under `romansworks/telnyx`. Before building, the release workflow SHALL pass the release gate (`make api-check-release`), the unit tests and the generated-docs check, and SHALL fail without publishing anything if any of them fails.

#### Scenario: Tag pushed
- **WHEN** `v1.2.0` is tagged
- **THEN** a GitHub release with signed checksums is created and the registries pick it up

#### Scenario: Release gate fails
- **WHEN** a `v*` tag is pushed while an operation in `catalog/coverage.yaml` is unreviewed
- **THEN** the release workflow fails before GoReleaser runs and no GitHub release is created

## ADDED Requirements

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
