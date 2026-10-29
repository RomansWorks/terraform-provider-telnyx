---
name: telnyx-sync
description: Bring the provider up to date with the latest Telnyx API. Use when asked to "run the Telnyx sync", update to the latest Telnyx API, or work on a sync/telnyx-api pull request.
---

# Telnyx API sync

Takes the provider from "Telnyx published a new spec" to a reviewed pull request:
fetch the spec, diff it, write an OpenSpec change, implement it, test it, archive it.
Read `AGENTS.md` first if you have not in this session.

## 1. Get the new spec

- If a `sync/telnyx-api` branch exists on origin (the daily workflow opened it),
  check it out: the new pin and catalog are already committed there.
- Otherwise, from an up-to-date `main`, create a branch `sync/telnyx-<YYYY-MM-DD>`
  and run `make sync`.
- If `make upstream-fetch` reports the spec is up to date and there is no sync
  branch, stop and say so. There is nothing to do.

## 2. Create the sync change

- Name it `sync-telnyx-<YYYY-MM-DD>-<sha7>`, where sha7 is the first 7
  characters of `commit` in `upstream/telnyx/lock.json`.
- `openspec new change <name> --schema telnyx-sync`
- `make api-diff CHANGE=<name> BASE=origin/main`
- Read `openspec/changes/<name>/api-diff.md` in full.

## 3. Write the plan (OpenSpec propose)

Follow the `openspec-propose` skill for the change created above, with these
inputs:

- For each changed or added candidate resource, read only its product area:
  `upstream/telnyx/by-tag/<tag-slug>.json`, its entry in `catalog/resources.yaml`,
  and the existing spec with `openspec show <area> --type spec`.
- Every item in api-diff.md must be accounted for in the proposal: implemented
  in this change, deferred under a "Deferred" heading with a reason, or
  classified in `catalog/coverage.yaml` (for example `out_of_scope` with a note).
- Removed or retyped fields are **BREAKING** for users. Plan a deprecation, never
  an immediate removal (see `openspec/specs/release/spec.md`).
- Run `openspec validate <name> --strict` and fix what it reports.

**Stop for review** and summarise the proposal for the user when it contains a
BREAKING item, removes a resource, or reclassifies an implemented resource.
Otherwise continue.

## 4. Implement (OpenSpec apply)

Follow the `openspec-apply-change` skill for the change. For each task:

- Change generator overrides or hand-written code, then `make generate`. Never edit
  files marked as generated.
- Update the operation's entry in `catalog/coverage.yaml` and set `reviewed: true`.
- Add or update unit tests (recorded responses), the acceptance test with an import
  step, the example under `examples/`, and a changie entry (`changie new`).
- Run `make check` until it passes.
- If `TELNYX_API_KEY` is set, run acceptance tests for the touched resources:
  `TF_ACC=1 go test ./internal/... -run '<TestAcc pattern>'`. If it is not set,
  say that acceptance tests were not run.

## 5. Finish

- `openspec archive <name> --yes` merges the spec deltas into `openspec/specs/`.
- `make check` once more.
- Commit, push and open (or update) the pull request. In its description, list
  what was implemented, what was deferred and why, and which tests ran.
