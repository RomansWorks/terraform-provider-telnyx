# Design

Wave 4 is generated from overrides. Two read hooks fill attributes the API answers under another name, and the generator learned two general rules that later waves can reuse.

## Forced deletes

Several Telnyx deletes refuse objects in use unless the request carries flags: `DELETE /email_domains/{id}` needs `force=true` for a verified domain, and the IP pool and dedicated IP deletes need `force=true&drain=true` for the last IP of a pool. A Terraform user cannot pass query parameters at destroy time, so the override names them:

```yaml
type: telnyx_email_ip_pool
delete: DELETE /email/ip_pools/{id}
force_destroy: [force, drain]
```

The generator checks that each name is a boolean query parameter of the delete operation and adds a local, optional `force_destroy` attribute, the name Terraform users know from other providers. Destroy reads it from state, like `release_on_destroy`, so it has to be applied before destroying. Other delete query parameters (`reason`) are not sent.

## Lists with two page shapes

`GET /email_blocks` answers with `oneOf` a page-number shape and a cursor shape; the variants differ only in `meta`. Merging the variants fails, so the list was a generation error. When a list response has no `data` array, the generator now takes it from the first `oneOf` or `anyOf` variant that has one. Paging is derived from the parameters as before, so `page[number]` wins.

## Lowercase addresses

Telnyx trims and lowercases email addresses, inbox usernames and sender filter entries before storing them. A configured `Someone@Example.com` would read back as `someone@example.com`: a replacement on every plan for a field that forces replacement, and for `telnyx_email_unsubscribe_group_suppression`, which is found in its group's list by address, a failed read right after create. `attributes.<f>.lowercase: true` adds a validator that rejects values with capitals or surrounding spaces (for a list or set, each element), so the mistake shows at plan time. Lowercasing silently would hide what is stored; a case-insensitive type would be more code than the problem needs.

## Soft deletes

`DELETE /email_blocks/{id}` sets a suppression's status to `removed`, and a released dedicated IP stays readable as `released` until its quarantine ends. A read that still answers 200 would keep such objects in state forever. A read hook answers with a not-found error for such a status; resource read then removes the object from state, as for a 404, and the next plan creates it again.

## Local-only changes

A resource without an update operation forces replacement for every API attribute, but `force_destroy` (and `release_on_destroy`) are local and change in place. Main already stores such an update without a request (wave 6); wave 4 adds the requirement and a test that `force_destroy` set on a membership is stored and used on destroy.

## Resources

| Resource | How | Notes |
|---|---|---|
| `telnyx_email_domain` | overrides, read hook | `force_destroy` sends `force=true`. `inbound_enabled` is answered as `inbound.enabled`; the hook copies it back. |
| `telnyx_email_domain_webhook` | overrides | Child of a domain, imports as `<domain_id>/<id>`. `events` is a set. |
| `telnyx_email_inbox` | overrides, read hook | No update: changes replace. The API answers with `address` only; the hook puts its local part into `username`, which is computed when not set. Delete is a soft delete and the address stays reserved. |
| `telnyx_email_inbox_sender_filters` | overrides | Settings of an inbox (`id_from_path: inbox_id`), replaced as a whole with `PUT`. Both lists are sets of lowercase entries. They are optional and computed, so removing one from configuration keeps the current list; `[]` clears it. The incremental `POST` and `DELETE` stay out of scope. No delete: destroy leaves the lists and warns. |
| `telnyx_email_dedicated_ip` | overrides | Purchase and release. `PATCH` is one imperative change per request (move pool, pause warm-up), so it is out of scope: `pool_id` is computed, membership is `telnyx_email_ip_pool_ip`, warm-up the four actions. `force_destroy` releases the last IP of a pool. A `released` IP is gone. |
| `telnyx_email_ip_pool` | overrides | `force_destroy` deletes a pool that still has IPs. |
| `telnyx_email_ip_pool_ip` | overrides | Read from the pool's `ips` field, keyed by the IP's id. Imports as `<pool_id>/<dedicated_ip_id>`. No update: a new weight replaces it, which for the last IP of a pool needs `force_destroy`. `PATCH /email/dedicated_ips/{id}` with the current pool could change the weight in place, but its `{id}` is the IP while this resource's `id` path parameter is the pool; mapping it is left for a later change. |
| `telnyx_email_suppression` | overrides, read hook | No update: changes replace. `to` and `from` are lowercase. A `removed` suppression is gone. |
| `telnyx_email_template` | overrides | `variables` stays a list: in explicit mode the API keeps its order. |
| `telnyx_email_unsubscribe_group` | overrides | Recipients who unsubscribe add suppressions outside Terraform, and a group with active suppressions is deleted only with `force=true`: `force_destroy`. Its `force` parameter takes a boolean or the strings `"true"`/`"false"`; the generator accepts a union with a boolean variant. |
| `telnyx_email_unsubscribe_group_suppression` | overrides | Identified by the recipient address (`to`), which the delete path takes as `{email}`; read from the group's suppression list. Imports as `<group_id>/<to>`. |

The data sources are the lists of these resources, `telnyx_email_domain_dns_records` and `telnyx_email_domain_health`. The actions are the four dedicated IP warm-up steps, domain verification and DKIM rotation.

## Known limits

- `telnyx_email_ip_pool.domain_id` and `telnyx_email_domain.dmarc_policy` are optional and computed: removing them from configuration keeps the current value. Clearing them takes an explicit value; this is the generator's general rule for optional and computed fields.

## What is not covered

- `POST /email_blocks/import` uploads a CSV as multipart and returns an import job. Actions cannot upload files or return the job id, and a one-off bulk import is not configuration: `out_of_scope`, with the job status read `GET /email_blocks/import/{id}`.
- `PATCH /email/dedicated_ips/{id}`: see above.
- `POST /email/dedicated_ips` requires an `Idempotency-Key` header. Wave 5 adds it to every write in `internal/client`; until that merges, purchasing a dedicated IP fails with `400`. The acceptance run in step 4 checks it.

## Testing

The conformance test runs every wave 4 resource through create, an empty plan, its data source, import and destroy. Its fake API now answers a `read_from` with `field` from the parent object, and a member added by an id in the request (`dedicated_ip_id`) is keyed by that id. Engine tests cover the `force_destroy` query, setting it on a resource without update, a not-found read hook and the lowercase validator; codegen tests the override key and union lists, and unit tests the two read hooks. Acceptance tests against a real account come in step 4.
