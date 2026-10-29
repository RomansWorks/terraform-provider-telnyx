# Design

## Context

The spec is OpenAPI 3.1, about 8 MB, with 1,505 operations over 230 tags. It has 92 operationIds that are shared by aliased paths, so operations are keyed by `METHOD /path`. Updates use POST, PATCH or PUT depending on the area, and some objects cannot be deleted.

## Goals / Non-Goals

**Goals:** a deterministic, offline pipeline from a pinned spec to agent-sized views; a CI-enforced definition of coverage; a provider core that later resources share.

**Non-Goals:** the resource code generator and any resource. Those follow in a separate change so the generator design gets its own review.

## Decisions

- **Own HTTP core instead of telnyx-go.** The SDK trails the spec by about 95 operations and would couple our releases to its releases. The core is small (auth, retries, errors) and request and response types will be generated from the same pinned spec.
- **Pin by commit and sha256.** `upstream/telnyx/lock.json` records the commit; CI fails if the spec file no longer matches its hash, so nobody edits it by hand.
- **Generated views are committed.** The per-tag split and catalog are checked in so reviewers see API changes as diffs, and so agents can read them without running tools.
- **Our own diff instead of oasdiff.** oasdiff 1.35 cannot load this OpenAPI 3.1 spec (it rejects string-valued `maximum`), and its 3.1 support is still in beta. `apisync diff` reports changes in the provider's own terms (operations and candidate resources). Revisit oasdiff once 3.1 support is stable.
- **Retries.** 429 is retried for every method because the request was not processed. Transport errors and 500/502/503/504 are retried only for GET, PUT, DELETE and HEAD, so a POST that may have succeeded is never sent twice. Backoff is exponential with full jitter, capped at 30 seconds, and honours `Retry-After`.
- **Coverage heuristic, human review.** `make catalog` fills new entries with a suggested bucket and `reviewed: false`; it never changes an existing entry. Review happens inside OpenSpec changes.

## Risks / Trade-offs

- [The catalog heuristic misses resources whose shape is unusual] → every operation still appears in coverage.yaml, so an unusual resource shows up as an unreviewed `action` or `data_source` entry and is caught in review.
- [Per-tag files for large areas such as Assistants are still several hundred kilobytes] → later work can add per-resource views generated the same way.
- [The pinned spec adds about 8 MB per changed version to git history] → acceptable for now; can move to Git LFS if the repository grows too large.
