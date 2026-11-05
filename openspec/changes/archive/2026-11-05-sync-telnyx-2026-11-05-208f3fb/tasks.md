# Tasks

## 1. Classify

- [x] 1.1 Classify the 38 Agent Gateway operations (resources and data sources for wave 5, spend logs out of scope)
- [x] 1.2 Classify the two operations left unreviewed by the `8938d09` pin
- [x] 1.3 Confirm the 19 changed operations need no provider change (`make generate` leaves the tree unchanged)
- [x] 1.4 Add the Agent Gateway resources and data sources to the wave 5 plan

## 2. Check

- [x] 2.1 `make api-check-release` passes (no unreviewed operations)
- [x] 2.2 `make check` passes
- [x] 2.3 Acceptance tests: not run, no provider code changed and `TELNYX_API_KEY` is not set
