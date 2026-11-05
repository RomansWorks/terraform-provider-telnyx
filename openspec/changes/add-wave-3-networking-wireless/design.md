# Design

Wave 3 is generated from overrides. Every resource is "generated with overrides"; one has a read hook (SIM card orders). The generator learned two general rules.

## Values only create returns

`POST /tunnels` answers with `tunnel_token`, which `GET /tunnels/{id}` never lists: Telnyx keeps only a hash. `POST /wireguard_peers` without a `public_key` generates a key pair and answers with `private_key` once; the read schema lists the field, but Telnyx does not store it.

```yaml
attributes:
  tunnel_token: {create_only: true}
```

`create_only` makes the attribute computed and sensitive. When the read schema lacks the field, it is taken from the create response schema. The engine takes the value from the create response and carries it through the read that follows create; on refresh it always keeps the value in state, so a refresh never clears it. Updates keep it too, because computed attributes use the state value when the plan does not know it. An attribute that a request takes cannot be `create_only`: the user's value would be the truth. Data sources and list data sources leave create-only attributes out, because a read never has them.

An import has no create response, so these values are null after import. That is the only way to see them in Terraform; users who imported a tunnel rotate its token outside Terraform.

## Read-only request objects

`PATCH /sim_cards/{id}` lists `status`, an object whose `value` and `reason` are each `readOnly`. The generator treated `status` as configurable because the object itself is not marked. A request object whose fields are all read-only is now dropped from the request, so `status` is computed. Regenerating changed no existing resource.

## Resource shapes

- `telnyx_sim_card`: `PATCH /sim_cards/{id}` as create, like `telnyx_phone_number`, with `sim_card_id` from the path and `release_on_destroy`. Delete decommissions the SIM card, which stops billing and cannot be undone, so it runs only when `release_on_destroy = true`.
- `telnyx_sim_card_group`: the private wireless gateway and wireless blocklist are set and removed through `updates` with their `set_*` and `remove_*` actions.
- `telnyx_network_default_gateway`: one per network, identified by `network_id` (`id_from_path`). `GET /networks/{id}/default_gateway` answers with a list, so it is read with `read_from` keyed by `network_id`.
- `telnyx_access_ip_range`: there is no single-range read; it is read from `GET /access_ip_ranges` keyed by `id`.
- `telnyx_sim_card_order`: orders cannot be cancelled; destroy removes the order from state and warns. Create takes `address_id`, but the response holds the address in `order_address`. A read hook copies `order_address.id` to `address_id`, so an imported order matches its configuration instead of planning a replacement, which would place a new, billable order.
- `telnyx_wireguard_peer`: a new `public_key` replaces the peer, so a private key Telnyx generated never stays in state next to a key the user supplied.
- `telnyx_sim_card`: `pin_puk_codes` is left out. Telnyx returns it only when a read sends `include_pin_puk_codes=true`, which the engine does not send, so it would always be null.
- Resources without an update (`telnyx_global_ip`, `telnyx_network_route`, WireGuard interfaces, gateways, tunnels, health checks, access IPs) replace on every change.

## Secrets and sets

- Sensitive: `telnyx_virtual_cross_connect.primary_bgp_key` and `secondary_bgp_key`. Create-only: `telnyx_tunnel.tunnel_token`, `telnyx_wireguard_peer.private_key`.
- Write-only: `telnyx_tailscale_subnet_router.auth_key_wo`. The spec marks the field `writeOnly` and never returns it; a replacement is accepted only before the router enrolls, so `auth_key_wo_version` sends it again through `PATCH`.
- Sets: SIM card `tags` and `authorized_imeis`, wireless blocklist `values`, traffic policy profile `services`, `ip_ranges` and `domains`, Tailscale `advertised_routes` (the spec says `uniqueItems`).

## Actions and their limits

Terraform actions return no values.
- `telnyx_tunnel_rotate_token` invalidates the token in state without showing the new one. It is for revoking a leaked token; to get a working token in state, replace the tunnel.
- `telnyx_esim_purchase` and `telnyx_sim_card_register` do not return the SIM cards. Users read them with `telnyx_sim_cards` and adopt them with `telnyx_sim_card`.
- SIM card enable, disable, standby, voice and public IP actions start asynchronous SIM card actions; `telnyx_sim_card_action(s)` read their progress.

## Data sources

`telnyx_wireless_blocklist_values` takes the required `type` query parameter. Its `data` is a JSON list, because each type answers with a different element shape. `telnyx_sim_card_device_details` and `telnyx_sim_card_public_ip` read one object for a SIM card `id`.

## Ownership

`telnyx_sim_card` is in wave 3, so the SIM card operations the scaffolding grouped with wave 6 (its actions, action data sources, device details, public IP and activation code) and the unassigned bulk SIM card actions and wireless regions are built here. The wave 6 thread agreed and does not build them.

## Asynchronous provisioning

Private wireless gateways, public internet gateways, WireGuard interfaces, virtual cross connects and Global IP health checks report a `status` and provision in the background; SIM card group `set_*` actions are asynchronous too. The engine does not wait for them. Without a real account it is unknown whether dependent creates (a peer on a provisioning interface, a group whose gateway is still provisioning) fail; the acceptance run in step 4 checks this and adds wait hooks where needed.

## What is not covered

Nothing in wave 3 moves to `out_of_scope`. The wireless reports and Global IP usage and latency reports stay out of scope as triaged.

## Testing

The conformance test runs all 23 resources through create, an empty plan, the data source, import and destroy against the descriptor-driven fake API; none is skipped. Unit tests cover `create_only` in codegen and in the engine, and read-only request objects. Acceptance tests against a real account come in step 4.
