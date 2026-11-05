# wireless Specification

## Purpose
Manages Telnyx wireless: SIM cards and eSIMs, SIM card groups and orders, data usage notifications, wireless blocklists and traffic policy profiles.

## Requirements

### Requirement: SIM cards
The provider SHALL manage `telnyx_sim_card`, a SIM card already on the account, identified and imported by `sim_card_id`. Create SHALL apply the configured settings with `PATCH /sim_cards/{id}`. Destroy SHALL remove the SIM card from state and keep it in Telnyx unless `release_on_destroy = true`, in which case it SHALL send `DELETE /sim_cards/{id}`, which decommissions the card.

#### Scenario: Adopt and tag a SIM card
- **WHEN** `telnyx_sim_card` with `sim_card_id` and `tags = ["tracker"]` is applied
- **THEN** the provider sends `PATCH /sim_cards/{id}` with the tags, and the next plan is empty

#### Scenario: Move to another group
- **WHEN** `sim_card_group_id` changes
- **THEN** the provider sends `PATCH /sim_cards/{id}` and the SIM card is not replaced

#### Scenario: Destroy keeps the card
- **WHEN** a SIM card without `release_on_destroy` is destroyed
- **THEN** no request is sent and it is removed from state with a warning

#### Scenario: Import a SIM card
- **WHEN** `terraform import telnyx_sim_card.x <sim_card_id>` runs
- **THEN** the SIM card is read and a matching configuration has an empty plan

### Requirement: SIM card attributes
A SIM card's `tags` and `authorized_imeis` SHALL be sets, and `status` SHALL be computed: it changes only through SIM card actions. PIN and PUK codes SHALL not be exposed, because reads do not request them.

#### Scenario: Reordered tags
- **WHEN** the API returns a SIM card's tags in another order
- **THEN** the plan is empty

### Requirement: SIM card groups
The provider SHALL manage `telnyx_sim_card_group`. Its `private_wireless_gateway_id` and `wireless_blocklist_id` SHALL be set and removed through the group's `set_*` and `remove_*` actions, without replacing the group.

#### Scenario: Remove the blocklist
- **WHEN** `wireless_blocklist_id` is removed from configuration
- **THEN** the provider sends `POST /sim_card_groups/{id}/actions/remove_wireless_blocklist`

### Requirement: SIM card orders
The provider SHALL manage `telnyx_sim_card_order`. Orders cannot be cancelled: every change SHALL replace the order, and destroy SHALL remove it from state and warn.

#### Scenario: Destroy an order
- **WHEN** a SIM card order is destroyed
- **THEN** no request is sent and the order is removed from state with a warning

### Requirement: Imported SIM card orders
A SIM card order's `address_id` SHALL be read from the response's `order_address.id`, so that importing an order does not plan a replacement.

#### Scenario: Import an order
- **WHEN** `terraform import telnyx_sim_card_order.x <id>` runs for an order shipped to address `a1`
- **THEN** `address_id` is `a1`, and a configuration with `address_id = "a1"` has an empty plan

### Requirement: Wireless policies
The provider SHALL manage `telnyx_wireless_blocklist`, `telnyx_traffic_policy_profile` and `telnyx_sim_card_data_usage_notification`. A blocklist's `values`, and a traffic policy profile's `services`, `ip_ranges` and `domains`, SHALL be sets. A blocklist's `type` SHALL force replacement.

#### Scenario: Reordered blocklist values
- **WHEN** the API returns a blocklist's values in another order
- **THEN** the plan is empty

### Requirement: SIM card actions
The provider SHALL offer actions to purchase eSIMs and register SIM cards, to enable, disable and put on standby a SIM card, to enable and disable voice for one SIM card or a group, and to set and remove public IPs.

#### Scenario: Enable a SIM card
- **WHEN** `action "telnyx_sim_card_enable"` is invoked with `sim_card_id = "s1"`
- **THEN** the provider sends `POST /sim_cards/s1/actions/enable` once

### Requirement: Wireless lookups
The provider SHALL offer data sources for SIM card, SIM card group and bulk SIM card actions, a SIM card's device details and public IP, wireless blocklist values by `type`, wireless regions and traffic policy profile services.

#### Scenario: Blocklist values of a type
- **WHEN** `data "telnyx_wireless_blocklist_values"` is read with `type = "country"`
- **THEN** the provider sends `GET /wireless_blocklist_values?type=country`, and `data` holds the list as JSON

### Requirement: eSIM activation code
The provider SHALL offer the ephemeral `telnyx_sim_card_activation_code`, the activation code of an eSIM read with `GET /sim_cards/{id}/activation_code`. It SHALL never be stored in plan or state.

#### Scenario: Pass the code on
- **WHEN** a configuration passes `ephemeral.telnyx_sim_card_activation_code.c.activation_code` to a write-only argument
- **THEN** the code reaches it and is not in state
