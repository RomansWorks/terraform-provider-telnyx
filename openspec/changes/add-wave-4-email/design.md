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

## Resources

| Resource | How | Notes |
|---|---|---|
| `telnyx_email_domain` | overrides, read hook | `force_destroy` sends `force=true`. `inbound_enabled` is answered as `inbound.enabled`; the hook copies it back. |
| `telnyx_email_domain_webhook` | overrides | Child of a domain, imports as `<domain_id>/<id>`. `events` is a set. |
| `telnyx_email_inbox` | overrides, read hook | No update: changes replace. The API answers with `address` only; the hook puts its local part into `username`, which is computed when not set. Delete is a soft delete and the address stays reserved. |
| `telnyx_email_inbox_sender_filters` | overrides | Settings of an inbox (`id_from_path: inbox_id`), replaced as a whole with `PUT`. Both lists are sets. The incremental `POST` and `DELETE` stay out of scope. No delete: destroy leaves the lists and warns. |
| `telnyx_email_dedicated_ip` | overrides | Purchase and release. `PATCH` is one imperative change per request (move pool, pause warm-up), so it is out of scope: `pool_id` is computed, membership is `telnyx_email_ip_pool_ip`, warm-up the four actions. `force_destroy` releases the last IP of a pool. |
| `telnyx_email_ip_pool` | overrides | `force_destroy` deletes a pool that still has IPs. |
| `telnyx_email_ip_pool_ip` | overrides | Read from the pool's `ips` field, keyed by the IP's id. Imports as `<pool_id>/<dedicated_ip_id>`. No update: a new weight replaces it. |
| `telnyx_email_suppression` | overrides | No update: changes replace. |
| `telnyx_email_template` | overrides | `variables` stays a list: in explicit mode the API keeps its order. |
| `telnyx_email_unsubscribe_group` | overrides | |
| `telnyx_email_unsubscribe_group_suppression` | overrides | Identified by the recipient address (`to`), which the delete path takes as `{email}`; read from the group's suppression list. Imports as `<group_id>/<to>`. |

The data sources are the lists of these resources, `telnyx_email_domain_dns_records` and `telnyx_email_domain_health`. The actions are the four dedicated IP warm-up steps, domain verification and DKIM rotation.

## What is not covered

- `POST /email_blocks/import` uploads a CSV as multipart and returns an import job. Actions cannot upload files or return the job id, and a one-off bulk import is not configuration: `out_of_scope`, with the job status read `GET /email_blocks/import/{id}`.
- `PATCH /email/dedicated_ips/{id}`: see above.
- `POST /email/dedicated_ips` requires an `Idempotency-Key` header. Wave 5 adds it to every write in `internal/client`; until that merges, purchasing a dedicated IP fails with `400`. The acceptance run in step 4 checks it.

## Testing

The conformance test runs every wave 4 resource through create, an empty plan, its data source, import and destroy. Its fake API now answers a `read_from` with `field` from the parent object, and a member added by an id in the request (`dedicated_ip_id`) is keyed by that id. Engine tests cover the `force_destroy` query, codegen tests the override key and union lists, and unit tests the two read hooks. Acceptance tests against a real account come in step 4.
