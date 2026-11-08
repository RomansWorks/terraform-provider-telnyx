# Proposal

## Why

The provider is meant for Terraform and OpenTofu, but nothing tested it with OpenTofu, and the actions (Terraform 1.14) were never checked against an OpenTofu CLI. The examples were not validated either: one referred to an attribute that does not exist.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **OpenTofu in CI.** A job runs the unit tests against the fake API with OpenTofu 1.11 (the first release with ephemeral resources and write-only attributes) and the latest 1.13, and validates the examples with it.
- **Example validation.** `cmd/examplecheck` (`make examples-check`) validates every example against a fresh build with either CLI, without credentials. CI runs it with Terraform, and with OpenTofu for everything except actions.
- **Documented support.** A guide states what works in OpenTofu: resources, data sources, ephemeral resources and write-only attributes from 1.11; actions and list resources only in Terraform.
- **Example fix.** `telnyx_ai_collection_source` refers to the collection's `id`.

## Capabilities

### Modified Capabilities
- `release`: supported CLIs and their CI coverage.

## Impact

- `.github/workflows/test.yml`, `GNUmakefile`, `cmd/examplecheck`, `templates/guides/opentofu.md.tmpl`, one example.
