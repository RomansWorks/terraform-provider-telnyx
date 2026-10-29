# Spec Delta

## Purpose

Keeps a pinned, verifiable copy of the Telnyx OpenAPI spec and derives the views and reports that agents use to keep the provider in step with the API.

## ADDED Requirements

### Requirement: Pinned upstream spec
The repository SHALL hold the Telnyx spec from `team-telnyx/openapi` (`openapi/spec3.json`) at `upstream/telnyx/spec3.json`, with `upstream/telnyx/lock.json` recording the repository, ref, commit and sha256. `make check` SHALL fail when the file does not match the lock.

#### Scenario: Hand edit detected
- **WHEN** someone edits the pinned spec without updating it through `make upstream-fetch`
- **THEN** `make check` fails and says to run `make upstream-fetch`

#### Scenario: Fetch with no upstream change
- **WHEN** `make upstream-fetch` runs and the upstream commit equals the lock's commit
- **THEN** no file changes

### Requirement: Agent-sized views
`make catalog` SHALL write one self-contained OpenAPI document per tag to `upstream/telnyx/by-tag/<tag-slug>.json`, holding only that tag's operations and the components they reference, and SHALL write `catalog/resources.yaml` listing candidate resources with their lifecycle operations and a schema hash. Output SHALL be deterministic.

#### Scenario: One product area
- **WHEN** an agent needs the Assistants API
- **THEN** it reads `upstream/telnyx/by-tag/assistants.json`, which resolves every reference it contains

#### Scenario: Re-running with no change
- **WHEN** `make catalog` runs twice on the same spec
- **THEN** the second run changes no file

### Requirement: Operation coverage
`catalog/coverage.yaml` SHALL have exactly one entry for every operation in the pinned spec, keyed by `METHOD /path`, with a bucket of `resource`, `data_source`, `action`, `ephemeral` or `out_of_scope`. `resource` and `ephemeral` entries SHALL name a Terraform type; `out_of_scope` entries SHALL give a reason. `make check` SHALL fail on missing, stale or invalid entries.

#### Scenario: New upstream operation
- **WHEN** a sync adds an operation
- **THEN** `make catalog` adds an entry with a suggested bucket and `reviewed: false`, and leaves existing entries untouched

#### Scenario: Removed upstream operation
- **WHEN** a sync removes an operation
- **THEN** `make catalog` removes its entry and the API diff lists it

#### Scenario: Release readiness
- **WHEN** `apisync check -require-reviewed` runs while any entry has `reviewed: false`
- **THEN** it fails

### Requirement: API diff report
`make api-diff` SHALL compare the spec at the previously pinned commit with the current pinned spec and write a Markdown report listing operations added, removed, changed and newly deprecated, and candidate resources added, removed and changed.

#### Scenario: Start of a sync change
- **WHEN** a sync change is created with the telnyx-sync schema
- **THEN** its first artifact is the generated `api-diff.md`

### Requirement: Scheduled upstream check
A scheduled workflow SHALL run at least daily, fetch the upstream spec, and when the commit changed, open or update one pull request containing the new pin, regenerated views and the API diff report.

#### Scenario: Upstream changed overnight
- **WHEN** Telnyx publishes a new spec commit
- **THEN** a pull request titled with the new commit appears with the diff report in its description
