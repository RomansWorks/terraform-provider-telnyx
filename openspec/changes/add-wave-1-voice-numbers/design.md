# Design

Wave 1 is almost entirely overrides. The generator learned the few shapes it could not derive, each as a general rule rather than a per-resource exception, so later waves reuse them.

## Cursor paging

`page[after]` alone selects `PagingCursor`: the engine sends `page[size]`, then `page[after]` with the cursor from `meta.cursors.after` or `meta.end_cursor`, and stops when the cursor is empty or repeats. Many Telnyx list operations share one `page` schema that lists `number`, `size`, `after` and `before`; when `number` is offered, page numbers are used, because those operations report `meta.total_pages`. Other paging parameters (`offset`, `page_cursor`, TeXML's `PageToken`) are still errors unless `paging: none`.

## Once-per-account settings

```yaml
type: telnyx_inbound_voice_channel_settings
create: PATCH /inbound_channels
read: GET /inbound_channels
update: PATCH /inbound_channels
singleton: true
```

The paths have no parameters, so there is nothing to identify the object by. The id is the type name without `telnyx_`, put into every response by the engine. Destroy removes it from state, like other objects without delete.

## Settings of another object

`id_from_path` already existed for one path parameter. Import now takes every path parameter in order, parents first, so `telnyx_external_connection_phone_number_settings` imports as `<external_connection_id>/<phone_number_id>`.

## Envelopes

A 2xx response is read from `data` when `data` is an object, or when every other property is `meta` or `links`. Telnyx always wraps objects this way; some responses also repeat the id next to `data`.

## Plain-text ephemeral values

`POST /telephony_credentials/{id}/token` answers with a JWT as `text/plain`. The generator gives such an operation one attribute, `value`, which overrides can rename (`token`). The client now sends `Accept: application/json, text/plain;q=0.9`.

## Query lookups

`GET /sip_registration_status` takes `credential_type` (required), `connection_id` and `username` as query parameters. A data-source-only `read` now turns scalar string query parameters into lookup attributes: required when the spec says so, otherwise optional. When the response has a field of the same name, it fills the attribute when it is not set.

## Batch creates

`POST /porting_orders/{id}/additional_documents` takes and returns a list. `telnyx_porting_order_additional_document` is one document; a create hook sends it as a list of one and keeps the single element of the response.

## What is not covered

- `GET /porting_orders/{id}/loa_template` returns a PDF. It is not data Terraform can use: `out_of_scope`.
- `GET /texml/Accounts/{account_sid}/Queues` returns `queues` with Twilio-style paging. `out_of_scope` for now; `telnyx_texml_queue` reads one queue.
- Terraform actions return no values. `telnyx_porting_order_share` creates a share token that Terraform cannot show; users who need the token call the API.

## Testing

The conformance test builds a configuration from each resource's required attributes, runs it against a fake API that routes requests by the descriptor's operations, and checks an empty plan after apply, the data source, import and destroy. Acceptance tests against a real account come in step 4.
