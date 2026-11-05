# Design

Wave 5 is mostly overrides. The generator and engine learned the shapes they could not derive, each as a general rule or an opt-in override key, so later waves reuse them. Everything is generated from the spec and overrides; four resources also have hooks, named below with the reason.

## Idempotency keys

The AI gateway, CloudFS and others declare a required `Idempotency-Key` header; wave 2 (RCS agents) and wave 4 (dedicated IPs) need the same. Rather than teach the generator about headers, `internal/client` sends a random version 4 UUID as `Idempotency-Key` on every POST, PUT, PATCH and DELETE. A retry of the same request reuses the key, so a create that timed out and is retried cannot create the object twice where the API honours the header; operations that do not declare it ignore it. Reads carry none.

## Secrets shown once

`POST /llm_token_gateway/token_keys` returns the key's `token`, and `POST /storage/cloudfs` its `meta_token`; reads never return them. Attributes come from the read schema, so these fields were lost. `attributes.<field>.create_only: true` adds the field from any 2xx create response (the token key's `200` is an idempotent replay without the token, so all codes are searched). The attribute is computed and `CreateOnly`. Wave 3 built the same key for tunnel tokens and merged first, so this change uses its implementation (the engine keeps the create response's value through every read) and adds only the search over every 2xx create response.

## Raw bodies

KV keys are stored with `PUT /storage/kvs/{id}/keys/{key}` and an `application/octet-stream` body, and read with a GET that answers with the raw bytes. A request body of raw bytes alone is no longer a generation error; the operation is marked `Raw`, and the engine refuses to send it without a `Send` hook. A read that answers in plain text or raw bytes becomes one `value` attribute (as plain-text ephemeral values already did), so drift in the value is detected. The client's `Accept` header now also accepts `application/octet-stream`.

`telnyx_storage_kv_key` has a `Send` hook that sends `value` as the body and `ttl_secs` as a query parameter. `value` is sensitive and stored as text: binary values are out of scope. The key list pages with `cursor` and `limit`, which the engine does not drive, so `telnyx_storage_kv_keys` reads one page (`paging: none`); `prefix` and `limit` narrow it.

## Lists in `results`

The AI test framework (`/ai/tests/test_framework/v1/...`) answers lists Django-style: `count`, `next`, `previous` and `results`, paged with `page` and `page_size`. Wave 2 added `records` lists with `PagingPage`; this generalises it: `listItems` names the items field, `Op.Results` reads `results` and `count`, and paging stops at a short page or when `count` items are read.

## Typed query lookups

`GET /ai/assistants/{assistant_id}/versions/{version_id}` takes `include_mcp_servers`, a boolean. Data source lookups now accept boolean and integer query parameters and send them in their JSON form; other types are still generation errors.

## Unions per provider

`POST /voice_clones` and `POST /voice_clones/from_upload` take a `oneOf` with one object per provider and model. The variants have the same fields with different enums and required sets, and no `type` tag, so the generator made the body one JSON attribute with nothing configurable. `merge_variants: true` merges such variants into one object: fields of the same type, enums joined (dropped when a variant takes any value), and a field required only when every variant requires it. It is opt-in, because applied everywhere it would turn wave 1's `telnyx_custom_storage_credential` `configuration` (one object per storage backend) into a structured object, a breaking change outside this wave.

## Hooks

- **`telnyx_storage_kv_key`** (`Send`): raw body, above.
- **`telnyx_storage_bucket_ssl_certificate`** (`Send`): `PUT /storage/buckets/{bucketName}/ssl_certificate` takes the certificate and key as files of a multipart form. Users give the PEM text (`certificate_pem`, `private_key_pem`, the key sensitive); the hook uploads both on create and update. The API returns only the certificate's details, so the PEM text is not checked for drift.
- **`telnyx_voice_clone`** (`ModifyPlan`, `Create`): a clone is made from a voice design (`POST /voice_clones`, JSON) or from a recording (`POST /voice_clones/from_upload`, multipart). One resource covers both, because update (`PATCH /voice_clones/{id}`), delete and the list read are shared, and coverage gives every operation one Terraform type. The descriptor's create is the upload, whose fields are a superset; the hook posts JSON with the design fields when `voice_design_id` is set. Exactly one of `voice_design_id` and `audio_path` is set. Like avatars, the recording's SHA-256 is tracked so a new recording replaces the clone and a moved file does not.
- **`telnyx_ai_gateway_token_group` and `telnyx_ai_gateway_token_key`** (`Headers`, `ModifyPlan`): update and delete need `If-Match` with the quoted version from state, so a change made outside Terraform since the last refresh fails with 412 instead of being overwritten. Each update raises the version, so the plan marks `version` and `updated_at` unknown when anything else changes.

`Hooks.Send` replaces the request of a create or update; `Hooks.Headers` adds headers to update and delete through `client.WithHeader`.

## Membership resources

- `telnyx_ai_assistant_tag` adds one tag (`POST /ai/assistants/{assistant_id}/tags`) and is read from the assistant's `tags` (`read_from` with `field`). Its id is the tag. Using it together with `tags` on `telnyx_ai_assistant` makes the two fight; the description says so.
- `telnyx_ai_conversation_insight_group_assignment` is read from the group's `insights`; its id is the insight id. The insight's own fields are skipped: they belong to `telnyx_ai_conversation_insight`.
- `telnyx_ai_collection_source` and `telnyx_ai_memory_namespace` (identified by `name`, which the paths use) are read from their lists, as are `telnyx_integration_secret` and `telnyx_voice_clone`, which have no single-item read.

## Settings

- `telnyx_ai_collection` changes its retrieval settings with `PATCH /ai/collections/{uuid}/settings` as an extra update, so a separate `telnyx_ai_collection_settings` resource would only conflict with it; the read of the settings is covered by the collection's read. `sources` is computed: sources are managed with `telnyx_ai_collection_source`.
- `telnyx_ai_memory_namespace_settings` and `telnyx_compute_function_log_export` are settings of another object, identified by its path parameter. The log export's `headers` are sensitive and never returned.
- `telnyx_ai_assistant_canary_deploy` exists at most once per assistant and is identified by `assistant_id`.

## Objects without update or delete

`telnyx_storage_kv_namespace`, `telnyx_storage_sql_database`, `telnyx_storage_migration_source`, `telnyx_integration_secret` and `telnyx_ai_memory_namespace` have no update: every argument forces replacement. `telnyx_storage_migration` has no delete: destroy removes it from state and warns, and the `telnyx_storage_migration_stop` action stops it.

## Secrets

Integration secret `token` and `password` and the migration source's `provider_auth` are sensitive; the integration secret's are never returned and kept from configuration. The gateway token and CloudFS meta token are sensitive and create-only, above.

## Names

`provider` is reserved by Terraform: it is `voice_provider` on voice clones and designs and `source_provider` on migration sources and the assistant import action. `GET /media` is `telnyx_media_files` (the plural of media). The generated Go file for `telnyx_ai_assistant_test` is `ai_assistant_test_resource.go`.

## Actions that create objects

Actions return nothing to Terraform. `telnyx_ai_assistant_clone`, `telnyx_ai_mission_clone` and `telnyx_ai_assistant_import` stay actions as the triage decided: the copies are deliberate one-offs, and their descriptions say to import them to manage them. The test metric bulk create moves to `out_of_scope`: `telnyx_ai_test_metric` manages metrics one by one, and a bulk create would leave objects Terraform does not track.

## Agent Gateway

The 208f3fb sync added the Agent Gateway API and assigned it to this wave; its design records the resource plan, which this change follows.

- **`telnyx_agent_gateway`** has no update, so every argument forces replacement. It is created `pending`, and its sub-objects answer 409 until it is `ready`, so a `Wait` hook polls it every 10 seconds within the `create` timeout (10 minutes by default; the `timeouts` block raises it). `failed` and `deleting` end the wait with an error; the engine has already saved the id, so the gateway stays in state and the next apply replaces it.
- **Sub-objects** without a single-item read (APIs, guardrails, keys, MCP servers, teams, policy attachments) are read from their gateway's list with `read_from`. Policies have a read by name.
- **Secrets.** The key's value is only in the create response: `key` is `create_only`, as for the AI gateway's token keys. Responses show only the `type` of an API's or MCP server's outbound `auth`, and leave the provider secrets out of a guardrail's `settings`, so both are write-only (`auth_wo`, `settings_wo`) with a version to send them again. That gives up drift detection of guardrail settings, but a sensitive JSON value compared against a response without its secrets would plan a change on every run. `static_headers` are returned, and sensitive because they often carry credentials. An ephemeral key resource, as the sync design suggested, can follow; it would need the API to return the value from a read.
- **Tool permissions** are one grant per key or team that `PUT` replaces. There is no delete, and the engine would only drop such objects from state, so the engine gains a generic `Delete` hook that also covers objects without a delete operation (wave 3 added the hook for objects with one). The hook writes `{"servers": []}`, which grants none.
- Model lists, tool lists, forwarded headers, guardrail modes and a policy's guardrail additions and removals are sets.

## Review fixes

An independent review found these, all fixed here:
- KV namespaces and SQL databases are provisioned asynchronously and their keys and queries answer 409 until `provision_ok`, so both wait on create, like the Agent Gateway.
- A clone made from a voice design rejects `label`, `model_id` and `ref_text`, which the design create does not take (the API fills `label` from the design's prompt, which would replace the clone on every plan). Updates always send `name`, which the update request requires. A missing recording no longer fails the plan of an existing clone.
- List data sources name their parent path parameter as the resource's `path_params` do (`namespace_id`, `gateway_id`, `collection_id`), which also renames three lists from earlier waves whose parameter was a bare `id`; nothing is released yet.
- The AI gateway lists require `page[snapshot]` after page 1: the engine now sends back `meta.snapshot` and stops on `meta.has_more = false`. A short `results` page no longer ends a list whose `count` says more remain.
- `create_only` searches every 2xx code the operation lists. Text reads reject values that are not UTF-8.

Not changed: `voice_provider` accepts both `Telnyx` and `telnyx` as the spec's enum does; the acceptance run will show whether the API normalises the case.

## What is not covered

- `GET`, `POST`, `PUT`, `DELETE /ai/missions/{mission_id}/knowledge-bases`, `/mcp-servers` and `/tools` have no request or response schema in the spec. There is nothing to derive attributes from; they return when Telnyx documents them.
- `PUT` and `DELETE /ai/assistants/{assistant_id}/tools/{tool_id}`: the assistant's read does not return the attached tool ids, so an attachment cannot be read back. `tool_ids` on `telnyx_ai_assistant` sets them.
- `GET /ai/assistants/{assistant_id}/texml` returns a bare TeXML string.
- `DELETE /ai/conversations/{conversation_id}/conversations-insights/{insight_id}` deletes runtime conversation data.
- Storage buckets: the spec has no operation that creates one; they are created through Telnyx's S3-compatible API. Roman decided (2026-11-05) that the provider documents the AWS provider for buckets instead of hand-writing a resource over the S3 API: a guide and an example show `aws_s3_bucket` pointed at Telnyx's S3 endpoint, and this provider manages the bucket's SSL certificate.

## Testing

The conformance test runs every generated resource except six that the descriptor-driven fake cannot play: the assistant tag and insight assignment (read from a parent), the KV key (raw bodies), the bucket SSL certificate and voice clones (multipart), and the Agent Gateway (waits for ready). Each has its own test against a stub API in `internal/provider/wave5_test.go`, or in `internal/provider/agent_gateway_test.go` (waiting for ready, a failed start, and tool permissions destroyed by an empty grant). They also check the AI gateway's `Idempotency-Key` and `If-Match` handling and that the token survives refresh. The client, engine and generator features have unit tests. Acceptance tests against a real account come in step 4.
