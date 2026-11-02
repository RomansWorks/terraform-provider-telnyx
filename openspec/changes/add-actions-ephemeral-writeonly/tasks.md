## 1. Engine

- [x] 1.1 Action descriptor and framework action: config from path parameters and body, invoke sends once, errors fail the action; tests against a fake API
- [x] 1.2 Ephemeral descriptor and framework ephemeral resource: open reads, result is sensitive, nothing in state; tests
- [x] 1.3 Write-only attributes: sent on create from config, sent on update only when `_wo_version` changes, never in state; tests
- [x] 1.4 Release on destroy: delete only when true, warning otherwise; tests

## 2. Generator

- [x] 2.1 Override keys `kind: action`, `kind: ephemeral`, `op`, `open`, `write_only`, `release_on_destroy`, with strict validation
- [x] 2.2 Derive action configuration, ephemeral results, `_wo` and `_wo_version` attributes and `release_on_destroy`
- [x] 2.3 Fixtures, goldens and error tests for each new shape

## 3. Provider

- [x] 3.1 Register generated actions and ephemeral resources; generated test checks coverage buckets for them
- [x] 3.2 `make check` passes
