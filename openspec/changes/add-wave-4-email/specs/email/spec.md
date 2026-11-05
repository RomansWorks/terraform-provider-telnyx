## ADDED Requirements

### Requirement: Email domains
The provider SHALL manage `telnyx_email_domain` and its child `telnyx_email_domain_webhook`. Changing `domain` SHALL replace the domain; other settings SHALL update in place. `inbound_enabled` SHALL be read from the API's `inbound.enabled`, so it detects drift and is filled on import. `force_destroy = true` SHALL delete a verified domain with `force=true`. A webhook SHALL import as `<domain_id>/<id>`, and its `events` SHALL be a set.

#### Scenario: Create and read a domain
- **WHEN** a user applies `telnyx_email_domain` with `domain` and `inbound_enabled = true`
- **THEN** the provider sends `POST /email_domains`, and the next plan is empty although the API answers `inbound.enabled`

#### Scenario: Change tracking
- **WHEN** `tracking.open_tracking` changes
- **THEN** the provider sends `PATCH /email_domains/{id}` and keeps the domain

#### Scenario: Destroy a verified domain
- **WHEN** a domain with `force_destroy = true` in state is destroyed
- **THEN** the provider sends `DELETE /email_domains/{id}?force=true`

#### Scenario: Import a webhook
- **WHEN** `terraform import telnyx_email_domain_webhook.w d1/w1` runs
- **THEN** the webhook `w1` of domain `d1` is in state

### Requirement: Email inboxes
The provider SHALL manage `telnyx_email_inbox`; a change SHALL replace it, and destroy SHALL soft-delete it with `DELETE /email_inboxes/{id}`. `username` SHALL be read from the local part of the inbox's `address`, and SHALL be computed when not set.

#### Scenario: Generated username
- **WHEN** an inbox is created without `username` and the API answers with `address = "x7k2@inbound.example.com"`
- **THEN** `username` is `x7k2` and the next plan is empty

#### Scenario: Import an inbox
- **WHEN** an inbox is imported by its id and its configuration sets the same `username`
- **THEN** the plan is empty

### Requirement: Inbox sender filters
`telnyx_email_inbox_sender_filters` SHALL manage an inbox's allowlist and blocklist as sets, replaced as a whole with `PUT /email_inboxes/{inbox_id}/filters`, and SHALL import by `inbox_id`. The API cannot delete them: destroy SHALL remove them from state and warn.

#### Scenario: Change the blocklist
- **WHEN** an address is added to `blocklist`
- **THEN** the provider sends `PUT /email_inboxes/{inbox_id}/filters` with both lists

#### Scenario: Destroy sender filters
- **WHEN** `telnyx_email_inbox_sender_filters` is destroyed
- **THEN** no request is sent, it leaves state, and a warning says the filters are still on the inbox

### Requirement: Dedicated IPs and IP pools
The provider SHALL manage `telnyx_email_dedicated_ip` (purchase on create, release on destroy), `telnyx_email_ip_pool` and `telnyx_email_ip_pool_ip`. A dedicated IP SHALL have no update: `region` and `ip_version` SHALL force replacement and `pool_id` SHALL be computed. The pool, the dedicated IP and the membership SHALL offer `force_destroy`, sending `force=true&drain=true`. The provider SHALL offer the dedicated IP warm-up actions: start, pause, resume and complete.

#### Scenario: Start the warm-up
- **WHEN** `action "telnyx_email_dedicated_ip_start_warmup" "x" { config { id = "ip1" } }` is invoked
- **THEN** the provider sends `POST /email/dedicated_ips/ip1/actions/start_warmup` once

### Requirement: IP pool membership
`telnyx_email_ip_pool_ip` SHALL be read from its pool's `ips`, identified by `dedicated_ip_id`, and import as `<pool_id>/<dedicated_ip_id>`. A new `weight` SHALL replace it. Destroy SHALL take the IP out of the pool and keep the dedicated IP.

#### Scenario: Add an IP to a pool
- **WHEN** a user applies `telnyx_email_ip_pool_ip` with `pool_id` and `dedicated_ip_id`
- **THEN** the provider sends `POST /email/ip_pools/{id}/ips`, reads the IP from `GET /email/ip_pools/{id}`, and the next plan is empty

#### Scenario: IP removed outside Terraform
- **WHEN** the IP is no longer in its pool's `ips`
- **THEN** the next plan proposes to add it again

#### Scenario: Remove the last IP
- **WHEN** the last membership of a pool, with `force_destroy = true` in state, is destroyed
- **THEN** the provider sends `DELETE /email/ip_pools/{id}/ips/{ip_id}?drain=true&force=true` and keeps the dedicated IP

### Requirement: Templates, suppressions and unsubscribe groups
The provider SHALL manage `telnyx_email_template`, `telnyx_email_suppression`, `telnyx_email_unsubscribe_group` and `telnyx_email_unsubscribe_group_suppression`. Suppressions have no update: a change SHALL replace them. A group suppression SHALL be identified by its recipient address `to`, sent as the `{email}` path parameter on delete, read from the group's suppression list, and import as `<group_id>/<to>`.

#### Scenario: Update a template
- **WHEN** the `html_body` of a template changes
- **THEN** the provider sends `PATCH /email_templates/{id}` and keeps the template

#### Scenario: Unsubscribe a recipient
- **WHEN** a `telnyx_email_unsubscribe_group_suppression` with `group_id = "g1"` and `to = "a@example.com"` is destroyed
- **THEN** the provider sends `DELETE /email_unsubscribe_groups/g1/suppressions/a@example.com`

### Requirement: Email lookups and actions
The provider SHALL offer the lists of the email resources, `telnyx_email_domain_dns_records` and `telnyx_email_domain_health` as data sources, and `telnyx_email_domain_verify` and `telnyx_email_domain_rotate_dkim` as actions.

#### Scenario: DNS records to publish
- **WHEN** `data "telnyx_email_domain_dns_records" "d" { domain_id = telnyx_email_domain.d.id }` is read
- **THEN** the provider sends `GET /email_domains/{domain_id}/dns_records` and lists the records
