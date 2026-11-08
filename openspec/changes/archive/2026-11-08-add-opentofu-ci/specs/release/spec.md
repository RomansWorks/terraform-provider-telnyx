## ADDED Requirements

### Requirement: Supported CLIs
The provider SHALL work with Terraform 1.11 or later and OpenTofu 1.11 or later. Actions SHALL need Terraform 1.14 or later and are not available in OpenTofu, which has no `action` blocks; everything else, including ephemeral resources and write-only attributes, SHALL work in both. CI SHALL run the unit tests and validate the examples with both CLIs, without Telnyx credentials, and the documentation SHALL say which features each CLI supports.

#### Scenario: OpenTofu user
- **WHEN** a configuration uses resources, data sources, ephemeral resources or write-only attributes with OpenTofu 1.11
- **THEN** it validates, plans and applies as with Terraform

#### Scenario: Action in OpenTofu
- **WHEN** an OpenTofu configuration declares an `action` block
- **THEN** OpenTofu rejects the block, and the guide explains that actions need Terraform 1.14

#### Scenario: Broken example
- **WHEN** a pull request adds an example that does not validate with either CLI
- **THEN** CI fails
