# Proposal

## Why

"All resources present" needs a finish line. After Phase 1, 1,481 of the 1,505 operations in the pinned spec still had the bucket `make catalog` guessed from path shapes (`reviewed: false`), so nobody could say how many resources the provider must have or when it is complete. The guesses were also often wrong: runtime objects such as faxes and browser sessions were shown as resources, aliased paths would have produced duplicate resources (`telnyx_oauth_client_2`), and 32 real configuration objects had been missed (8 more were missing a create, delete or update).

This change covers the whole pinned spec, `team-telnyx/openapi@3d1ee2b`, in every product area.

## What Changes

- Every entry in `catalog/coverage.yaml` has been reviewed (`reviewed: true`) using the classification rules in `design.md`. Each entry now has a final bucket, a provisional Terraform name, and a note giving the reason.
- The results are 157 resources (7 already built), 297 data sources, 121 actions, 3 ephemeral resources, and 573 operations out of scope.
- `apisync check` also fails when a reviewed `data_source` or `action` entry names no Terraform type. It also fails when an `out_of_scope` entry names one. This way, every reviewed entry says what it becomes, and only that.
- The resource list and the six build waves are recorded in `design.md`. The engine features the waves need are listed there too.

No provider behaviour changes: nothing is added to or removed from the provider in this change.

## Capabilities

### Modified Capabilities
- `api-sync`: operation coverage names a type for every reviewed entry that is not out of scope, and follows written classification rules.

## Impact

- `catalog/coverage.yaml` is rewritten in full. `make catalog` still never changes existing entries.
- `internal/apisync/coverage.go`: the check is stricter for reviewed entries.
- New operations from later syncs are reviewed by hand in each sync change, following the same rules. `make catalog` still suggests names from path shapes.
- Names are provisional until a wave implements them. A wave may rename an entry and must record the rename in its own change. Nothing is released under these names yet, so a rename breaks no user.
