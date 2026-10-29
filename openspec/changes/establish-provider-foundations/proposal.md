# Proposal

## Why

The provider needs a foundation before any Telnyx resource can be built: a configured provider and HTTP core, a pinned copy of the Telnyx API contract that agents can read in pieces, a way to prove every API operation is accounted for, and release rules. Telnyx changes its spec most working days, so keeping up has to be routine work for a coding agent.

## What Changes

- Provider block `telnyx` with `api_key`, `base_url` and `max_retries`, each settable from environment variables.
- Shared HTTP core: bearer authentication, retries with backoff for rate limits and idempotent server errors, and error messages that carry Telnyx error codes and the request id.
- The Telnyx OpenAPI spec pinned to an upstream commit with a lock file, split per product area, and indexed as a resource catalog.
- `catalog/coverage.yaml` classifying all 1,505 operations, checked in CI.
- An API diff report between pinned spec versions, used to start each sync change.
- A daily job that opens a pull request when Telnyx publishes a new spec.
- Release rules: semantic versioning, deprecation before removal, signed releases to the Terraform and OpenTofu registries.

## Capabilities

### New Capabilities
- `provider-core`: provider configuration, authentication, retries and error reporting shared by every resource.
- `api-sync`: pinning, splitting, cataloguing, classifying and diffing the upstream Telnyx spec.
- `release`: versioning, deprecation, changelog and publishing rules.

### Modified Capabilities

## Impact

New Go module `github.com/RomansWorks/terraform-provider-telnyx`, the `apisync` tool, `upstream/` and `catalog/` data, CI workflows, and agent instructions. No resources yet; the generator and the first resources come in the next change.
