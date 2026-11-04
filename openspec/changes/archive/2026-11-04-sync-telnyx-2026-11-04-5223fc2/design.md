# Design

## Context

The new field `language` is a plain optional string in the assistant's voice settings. The generator derives it as optional and computed, so existing configurations get an empty plan.

## Goals / Non-Goals

**Goals:** bring `main` back to green, expose the new field, and classify the new operations.

**Non-Goals:** implementing `telnyx_user_profile` now. It is an account setting, so it belongs with the account wave, where it gets its own requirement and tests.

## Decisions

- No override for `language`. The spec lists the accepted values only in prose and they depend on the voice provider, so the provider does not validate them and the API's error is returned on apply.
- `telnyx_user_profile` stays in the `resource` bucket with `reviewed: true` and a note that it is built in wave 6.

## Risks / Trade-offs

- The user profile API is marked beta upstream and may change before wave 6.
