## MODIFIED Requirements

### Requirement: Scheduled upstream check
A scheduled workflow SHALL run at least daily, fetch the upstream spec, and when the commit changed, open or update one pull request containing the new pin, regenerated views, the API diff report, and the output of `make generate` (resource descriptors, docs and formatted examples). When `make generate` fails, the pull request SHALL still open, and its description SHALL say that generation failed.

#### Scenario: Upstream changed overnight
- **WHEN** Telnyx publishes a new spec commit
- **THEN** a pull request titled with the new commit appears with the diff report in its description

#### Scenario: Generated docs stay current
- **WHEN** the new spec changes an attribute description of an implemented resource
- **THEN** the sync pull request includes the regenerated descriptor and docs, so the `Generated docs` check passes

#### Scenario: Override names a removed operation
- **WHEN** `make generate` fails because an override uses an operation the new spec removed
- **THEN** the pull request opens with the pin and catalog, and its description says generation failed and must be fixed in the sync change
