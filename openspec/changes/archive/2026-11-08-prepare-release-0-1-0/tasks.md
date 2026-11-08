## 1. Pipeline

- [x] 1.1 Release workflow: verify the tag has `.changes/<version>.md`, run `make api-check-release`, unit tests and the generated-docs check before GoReleaser
- [x] 1.2 Pass `.changes/<version>.md` to GoReleaser as release notes
- [x] 1.3 `make release-snapshot`; drop the unused `main.commit` ldflag
- [x] 1.4 `goreleaser check` and a snapshot build pass locally

## 2. Docs

- [x] 2.1 README: coverage, requirements, OpenTofu, release status
- [x] 2.2 `RELEASING.md`: one-time maintainer setup and the release steps
- [x] 2.3 Every example validates against a local build (`terraform validate` with dev overrides)

## 3. Changelog

- [x] 3.1 `changie batch 0.1.0 && changie merge` after the in-flight PRs merge

## 4. Check

- [x] 4.1 `make check` passes
