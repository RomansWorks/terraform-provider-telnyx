## 1. Engine and generator

- [x] 1.1 `create_only` attribute override: computed, sensitive, taken from the create response schema when read lacks it; engine keeps the state value on read; tests
- [x] 1.2 A request object whose fields are all read-only is computed; test; no existing resource changes

## 2. Wave 3 overrides

- [x] 2.1 23 resources: `telnyx_access_ip_address`, `telnyx_access_ip_range`, `telnyx_fqdn`, `telnyx_ip`, `telnyx_global_ip`, `telnyx_global_ip_assignment`, `telnyx_global_ip_health_check`, `telnyx_network`, `telnyx_network_default_gateway`, `telnyx_network_route`, `telnyx_private_wireless_gateway`, `telnyx_public_internet_gateway`, `telnyx_tailscale_subnet_router`, `telnyx_tunnel`, `telnyx_virtual_cross_connect`, `telnyx_wireguard_interface`, `telnyx_wireguard_peer`, `telnyx_sim_card`, `telnyx_sim_card_data_usage_notification`, `telnyx_sim_card_group`, `telnyx_sim_card_order`, `telnyx_traffic_policy_profile`, `telnyx_wireless_blocklist`; their 22 list data sources
- [x] 2.2 17 read-only data sources: Global IP allowed ports, health check types and protocols, network coverage, network interfaces, virtual cross connect coverage, traffic policy profile services, wireless blocklist values, wireless regions, SIM card group action(s), SIM card action(s), bulk SIM card action(s), SIM card device details and public IP
- [x] 2.3 13 actions: `telnyx_esim_purchase`, `telnyx_tunnel_rotate_token`, `telnyx_sim_card_register`, SIM card bulk voice and public IP actions, SIM card enable, disable, standby, voice and public IP actions
- [x] 2.4 Ephemeral `telnyx_wireguard_peer_config` and `telnyx_sim_card_activation_code`
- [x] 2.5 Secrets sensitive, create-only or write-only; unordered lists as sets; real descriptions

## 3. Tests, docs and release notes

- [x] 3.1 Conformance test passes for every wave 3 resource, none skipped
- [x] 3.2 `make generate` docs; examples for the main wave 3 resources, an action, the ephemeral resources and two data sources; import examples for every resource
- [x] 3.3 Changelog entries
- [x] 3.4 `make check` passes
