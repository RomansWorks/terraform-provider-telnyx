# Tasks

## 1. Repository foundations

- [x] 1.1 Create the module from the Plugin Framework scaffold, rename it to github.com/RomansWorks/terraform-provider-telnyx, and verify `go build ./...` succeeds
- [x] 1.2 Initialise OpenSpec with the telnyx-sync schema and verify `openspec schema validate telnyx-sync` passes

## 2. Provider core

- [x] 2.1 Implement the provider block and environment fallbacks, verified by TestProviderSchemaValid
- [x] 2.2 Implement the HTTP core with auth, retries and error decoding, verified by the internal/client tests

## 3. API sync

- [x] 3.1 Pin the upstream spec with a lock file and verify `apisync check` rejects a hand-edited spec (TestVerifyLock)
- [x] 3.2 Generate the per-tag split and resource catalog, verified by TestSplitByTagIsSelfContained and TestDetectResources
- [x] 3.3 Generate coverage.yaml for all operations and verify `make check` passes
- [x] 3.4 Implement the API diff report, verified by TestCompare and a run against an older upstream commit
- [x] 3.5 Add the scheduled sync workflow and verify it parses with actionlint

## 4. Release and agent workflow

- [x] 4.1 Configure GoReleaser, the registry manifest and changie, and verify `changie batch` dry-run works
- [x] 4.2 Write AGENTS.md and the telnyx-sync skill describing the update loop, and verify each command it names exists in the Makefile

## Workflow follow-up

- Archive this change once the setup pull request is merged.
