# networking Specification

## Purpose
Manages Telnyx private networking: networks and their interfaces, Global IPs, WireGuard, tunnels, cross connects and IP access lists.

## Requirements

### Requirement: Networks and interfaces
The provider SHALL manage `telnyx_network`, `telnyx_wireguard_interface`, `telnyx_wireguard_peer`, `telnyx_private_wireless_gateway`, `telnyx_public_internet_gateway`, `telnyx_virtual_cross_connect`, `telnyx_tailscale_subnet_router` and `telnyx_network_route`. Objects without an update operation SHALL be replaced when a configured attribute changes.

#### Scenario: Create and read a network
- **WHEN** `telnyx_network.office` with `name = "office"` is applied
- **THEN** the provider sends `POST /networks`, and the next plan is empty

#### Scenario: Rename a network
- **WHEN** `name` changes
- **THEN** the provider sends `PATCH /networks/{id}` and the network is not replaced

#### Scenario: Change a WireGuard interface
- **WHEN** `region_code` of a WireGuard interface changes
- **THEN** the interface is replaced, because the API cannot update it

#### Scenario: Delete and import
- **WHEN** a WireGuard peer is destroyed
- **THEN** the provider sends `DELETE /wireguard_peers/{id}`
- **WHEN** `terraform import telnyx_wireguard_peer.x <id>` runs
- **THEN** the peer is read and a matching configuration has an empty plan

### Requirement: Networking secrets
A virtual cross connect's `primary_bgp_key` and `secondary_bgp_key` SHALL be sensitive. A Tailscale subnet router's auth key SHALL be write-only (`auth_key_wo`, sent again when `auth_key_wo_version` changes), and its `advertised_routes` SHALL be a set.

#### Scenario: Replace a Tailscale auth key
- **WHEN** `auth_key_wo` changes together with `auth_key_wo_version`
- **THEN** the provider sends `PATCH /tailscale_subnet_routers/{id}` with the new `auth_key`, and state holds no `auth_key_wo` value

### Requirement: WireGuard peer private key
When a WireGuard peer is created without `public_key`, its `private_key` SHALL hold the private key Telnyx generated, as a sensitive value that refreshes never clear. After import it SHALL be null. A change of `public_key` SHALL replace the peer.

#### Scenario: Generated key pair
- **WHEN** a peer without `public_key` is created and the response holds `private_key`
- **THEN** `private_key` is in state and stays there after the next refresh

### Requirement: Default gateway of a network
The provider SHALL manage `telnyx_network_default_gateway`, one per network, identified and imported by `network_id`. It SHALL be read from `GET /networks/{id}/default_gateway`, and a change of `wireguard_peer_id` SHALL replace it.

#### Scenario: Import the default gateway
- **WHEN** `terraform import telnyx_network_default_gateway.x net1` runs
- **THEN** the gateway of network `net1` is read and a matching configuration has an empty plan

### Requirement: Tunnels
The provider SHALL manage `telnyx_tunnel`. Its `tunnel_token` SHALL be sensitive, filled on create and kept in state on refresh, because only create returns it. The `telnyx_tunnel_rotate_token` action SHALL rotate the token; because actions return no values, the token in state stops working after it runs.

#### Scenario: Token kept on refresh
- **WHEN** a tunnel refreshes
- **THEN** `tunnel_token` keeps the value from create and the plan is empty

### Requirement: Global IPs
The provider SHALL manage `telnyx_global_ip`, `telnyx_global_ip_assignment` and `telnyx_global_ip_health_check`, and offer data sources for their allowed ports, health check types and protocols.

#### Scenario: Assign a Global IP
- **WHEN** `telnyx_global_ip_assignment` with a Global IP and a WireGuard peer is applied
- **THEN** the provider sends `POST /global_ip_assignments`, and changing `is_in_maintenance` updates it with `PATCH`

### Requirement: IP access and connection addresses
The provider SHALL manage `telnyx_access_ip_address` and `telnyx_access_ip_range`, which allow addresses to reach the API, and `telnyx_fqdn` and `telnyx_ip`, the addresses of FQDN and IP connections. An access IP range SHALL be read from `GET /access_ip_ranges`, because there is no single-range read.

#### Scenario: Range removed outside Terraform
- **WHEN** an access IP range is no longer in `GET /access_ip_ranges`
- **THEN** it is removed from state

### Requirement: Networking lookups
The provider SHALL offer `telnyx_network_interfaces` for a network's interfaces, `telnyx_network_coverage`, `telnyx_virtual_cross_connect_coverage`, the WireGuard peer configuration as the ephemeral `telnyx_wireguard_peer_config`, and a list data source for each networking resource with a list operation.

#### Scenario: Peer configuration
- **WHEN** `ephemeral.telnyx_wireguard_peer_config.p` is opened for a peer
- **THEN** the configuration is read from `GET /wireguard_peers/{id}/config` and is not stored in state
