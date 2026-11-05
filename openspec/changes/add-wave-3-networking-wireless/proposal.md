# Proposal

## Why

Step 3 of the plan builds the provider in waves by product area. Wave 3 is networking and wireless: private networks and their interfaces, Global IPs, WireGuard, tunnels, cross connects, access lists for the API, and SIM cards with their groups, orders and blocklists. The coverage review assigned 23 resources to it. Two of them return a secret only in the create response (a tunnel's token, a WireGuard peer's generated private key), which the generator could not keep in state, and a SIM card's update body lists a status object that the API never lets a request set.

This change covers no upstream range: it builds on the current pin.

## What Changes

- **Generator and engine.**
  - `attributes.<field>.create_only: true` makes a computed, sensitive attribute that only create fills. The field may be missing from the read response; reads never overwrite the value in state.
  - A request object whose fields are all read-only is not configurable: it is computed.
- **Wave 3 resources, data sources, actions and ephemeral resources**, generated from overrides. The list is in `tasks.md`.
- **`telnyx_sim_card`** adopts a SIM card already on the account. Destroy keeps it unless `release_on_destroy = true`, as decided for numbers and SIM cards.
- **SIM card operations** the coverage review placed with account work (wave 6) move here with the SIM card: its actions, action and device data sources, the eSIM activation code, bulk SIM card actions and wireless regions.
- **`telnyx_tailscale_subnet_router`** takes its Tailscale auth key write-only (`auth_key_wo`).
- **A conformance run** covers every new resource; none is skipped.

## Capabilities

### Modified Capabilities
- `codegen`: create-only values, read-only request objects.

### New Capabilities
- `networking`: networks and their interfaces, Global IPs, WireGuard, tunnels, cross connects, IP access lists, FQDNs and IPs of connections.
- `wireless`: SIM cards, SIM card groups, orders, data usage notifications, wireless blocklists, traffic policy profiles, eSIM purchase and registration.

## Impact

- `internal/engine`: `Attr.CreateOnly`.
- `internal/codegen`: the attribute override key `create_only`; read-only request objects.
- `internal/provider`: one hook (SIM card orders read `address_id` from `order_address`).
- `generator/overrides`: 55 new override files. `internal/provider/generated`: regenerated.
- `catalog/coverage.yaml`: unchanged; every wave 3 operation is implemented as triaged.
- Docs and examples for the new resources; changelog entries.
