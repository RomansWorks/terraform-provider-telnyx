# Design

## Context

The new field `extra_headers` is an object with string values (`additionalProperties: string`), so the generator derives it as a JSON string, like the assistant's other free-form maps. It is optional and computed, so existing configurations get an empty plan.

## Goals / Non-Goals

**Goals:** bring `main` back to green and expose the new field.

**Non-Goals:** the AI assistant version, clone and import operations, which belong to a later wave.

## Decisions

- The derived shape is correct; the only override is marking `extra_headers` sensitive in its three places. Authentication should go through the existing `*_ref` secret references, but nothing stops a user from putting an API key in a custom header, so the values are hidden from plan output.

## Risks / Trade-offs

- The API rejects reserved header names; the provider does not check them, and the API's error is returned on apply.
