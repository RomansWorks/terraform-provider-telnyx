## Why

The provider covers every Telnyx API operation it plans to (catalog/coverage.yaml:
0 unreviewed entries) and is ready for its first registry release, 0.1.0. The
release pipeline is still the HashiCorp scaffold: it does not run the release
gate, publishes no release notes, and nothing documents what the maintainer must
set up before the first tag. This change makes the pipeline safe to run and the
repository ready to publish. It covers no upstream API range and changes no
resource behaviour.

## What Changes

- The release workflow runs the release gate (`make api-check-release`), the unit
  tests and a generated-docs check before GoReleaser, and refuses a tag that has
  no collated changelog section.
- GitHub releases carry the changelog section for their version as release notes.
- `make release-snapshot` builds every platform locally without signing or
  publishing, so the GoReleaser config can be checked before a tag.
- The unreleased changelog entries are collated into `0.1.0`.
- README and `RELEASING.md` describe the current coverage, requirements
  (Terraform 1.14 for actions) and the one-time setup the maintainer must do:
  GPG key, GitHub secrets, Terraform Registry and OpenTofu registry.

No breaking changes.

## Impact

- `.github/workflows/release.yml`, `.goreleaser.yml`, `GNUmakefile`
- `CHANGELOG.md`, `.changes/`
- `README.md`, `RELEASING.md`
- Spec: `release`
