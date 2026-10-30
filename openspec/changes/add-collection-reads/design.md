# Design

All of this is generated with overrides. The new override keys describe what the spec cannot say: which operation is the list, where a child appears in its parent, and which endpoint sets which attributes. Everything else is derived.

## List data sources

```yaml
# in a resource override
list: GET /billing_groups            # adds data source telnyx_billing_groups
list_name: telnyx_billing_groups     # optional; default is the plural of type
```

```yaml
# a data-source-only override
type: telnyx_country_coverage
kind: data_source
description: ...
read: GET /country_coverage          # a single object, or
list: GET /phone_numbers/jobs        # a list
```

- **Filters.** Query parameters become optional top-level attributes, named in snake_case after the parameter. A `filter[...]` deepObject parameter (`filter[status]`) becomes `filter_status`. Nested filter keys are joined with `_`, and their `eq` suffix is dropped. Parameters for paging and sorting are not attributes (`page`, `page[...]`, `page_number`, `page_size`); `sort` is an optional attribute.
- **Items.** The response's `data` array items (or the items of a bare array response, which some AI endpoints return) become the computed nested list attribute named after the plural noun, for example `billing_groups`. Their attributes are derived from the list response, not the item read, because list responses sometimes return fewer fields.
- **Paging.** The engine supports the two styles that cover 120 of the 130 lists:
  - `page[number]`/`page[size]` with `meta.total_pages`
  - flat `page_number`/`page_size`

  `page[number]` and `page[size]` may also appear as two flat query parameters with those names; they page the same way. A list with neither is read once. Without `meta.total_pages`, a page shorter than the page size is the last one. The generator rejects any other paging parameter unless the override sets `paging: none` and a hook supplies the paging. Page size is the spec's maximum, or 250, whichever is smaller.
- **Limit.** An optional `max_items` stops fetching once enough items have been read, so large accounts don't page through every number they own.

## Read from a collection

```yaml
read_from:
  op: GET /integration_secrets     # the list, or the parent read
  field: ""                        # "" = the response's data array; else a field of the object, such as tags
  key: id                          # element field matched against the id; "" for plain values
```

- The engine treats a missing element as not found. Read then removes the resource from state, and create fails with "not found after create".
- List reads page through with the same paging as list data sources.
- For plain values (assistant tags), the element is the value itself. The id attribute holds that value.
- Import ids are unchanged: they still name the parent ids and the id.

## Extra update endpoints

```yaml
updates:
  - op: POST /sim_card_groups/{id}/actions/set_wireless_blocklist
    clear: POST /sim_card_groups/{id}/actions/remove_wireless_blocklist
```

- The op's request schema contributes attributes. They are Optional and never Computed: removing one from configuration must plan it as null, so that the clear op runs. Read still fills them, so a change made outside Terraform shows as drift.
- On create, after the main create, the engine sends each extra op whose attributes are set.
- On update, the engine sends each op whose attributes changed. When all of an op's attributes are removed and a `clear` op exists, it sends the clear op instead.
- An attribute that only extra updates set does not force replacement.
- The main update never sends these attributes.

## Alternatives considered

- **Terraform list resources** (`list` blocks, Terraform 1.14) instead of plural data sources. Rejected for now: OpenTofu does not support them, and plural data sources work in every version. They can be added later from the same `List` descriptor.
- **One hook per resource for read-from-list.** Rejected: 19 resources would repeat the same loop, and each would need its own tests.
