# Design

## How the review was done

A one-off script applied the rules from the `api-sync` delta to every entry. It worked from three inputs: the base `catalog/coverage.yaml`, `catalog/resources.yaml`, and the pinned spec. Then it took a hand-written decision table, which a person reviewed product area by product area. The table has four parts:

- Candidates renamed to the names Telnyx uses (`telnyx_dir` became `telnyx_display_identity_record`; `telnyx_rc_agent` became `telnyx_rcs_agent`).
- Candidates dropped as runtime.
- Resources the path heuristics missed.
- A few forced entries.

The script is not kept. `coverage.yaml` and its notes are the record, and later syncs classify new operations by the same rules.

Each note says why the entry got its bucket:

| Note | Meaning |
|---|---|
| `create of telnyx_x (catalog collection)` | Lifecycle operation of a resource that `make catalog` detected |
| `create of telnyx_x (triage)` | Lifecycle operation of a resource added in this review |
| `list of telnyx_x` | Plural data source for the resource |
| `alias of METHOD /path` | Duplicate path; the canonical one carries the entry |
| `alternative update; …` | The resource uses the other update method |
| `runtime: …`, `reporting` | Out of scope, with the reason |
| `secret value: never stored in state` | Ephemeral |
| `lookup`, `reference data` | Data source |
| `one-shot operation` | Action |

## Naming

- **Resources** use the names Telnyx shows in its portal, not path fragments. Settings that belong to another object end in `_settings`, for example `telnyx_phone_number_messaging_settings`.
- **Plural data sources** are the resource name in the plural (`telnyx_billing_groups`). Other data sources are named from their path.
- **Actions** are named `telnyx_<noun>_<verb>`, for example `telnyx_sim_card_enable` or `telnyx_requirement_group_submit_for_approval`.

All names are provisional until their wave ships them.

## Decisions worth knowing

- **Dropped as runtime, although `make catalog` saw a create/delete lifecycle:**
  - browser sessions, AI conversations and clusters, mission runs and plan steps
  - faxes, messages, email messages and drafts
  - meeting sessions, room compositions, assistant scheduled events
  - queue calls, conference participants, TeXML calls and participants
  - OAuth grants
  - the legacy CDR, MDR, speech-to-text, telco and wireless report requests (`reporting`)
- **Added as resources:**
  - access IP ranges
  - AI assistant tags, tool attachments and canary deploys
  - AI collection sources, memory namespaces, insight group assignments
  - email IP pool IPs, unsubscribe-group suppressions, inbox filters
  - integration secrets
  - porting order documents, associated numbers, number blocks and extensions
  - RCS test devices, spend limits, voice clones, voice channel zones
  - the creates for custom storage credentials, Dialogflow connections, 10DLC campaigns (`POST /10dlc/campaignBuilder`) and display identity records
  - network default gateways, number reservations, voicemail, user addresses, enterprise reputation, verify message templates
  - the advanced, inexplicit, number block and SIM card orders, and storage migrations (like `telnyx_number_order`: created once, never deleted)
- **Ephemeral:**
  - `POST /telephony_credentials/{id}/token`
  - `GET /sim_cards/{id}/activation_code`
  - `GET /wireguard_peers/{id}/config`
- **Out of scope by tag**, unchanged from the foundations: the runtime and reporting tags in `internal/apisync/coverage.go`.

## Engine features the waves need

These belong to step 2 of the plan, and each gets its own change before the wave that needs it:

| Feature | Needed by |
|---|---|
| Plural (list) data sources with filters and pagination | 299 data source entries, about 120 of them lists |
| Read from the list operation when there is no item `GET` | 14 resources (integration secrets, spend limits, voice clones, porting order children, …) |
| Read from the parent object | 4 attachment resources (assistant tags, tool attachments, IP pool IPs, insight group assignments) |
| Terraform actions | 133 action entries; needs Terraform 1.14 or later |
| Ephemeral resources | 3 entries |
| Write-only secret attributes (`_wo`) | Integration secrets, telephony and push credentials, OAuth clients, storage credentials |
| Adopting an object the API cannot create | 24 resources (phone numbers, SIM cards, hosted numbers, settings objects) |
| No delete in the API: destroy removes from state and warns | 33 resources (already supported) |

## Waves

Each wave is one OpenSpec change and one PR, built in this order. A wave also ships the data sources, actions and ephemeral resources of its area once the engine supports them. Struck-through names are the Phase 1 pilots.

### Wave 1: Voice and numbers

47 resources, 70 data sources, 28 actions, 1 ephemeral resource.

`telnyx_address`, `telnyx_advanced_number_order`, `telnyx_billing_group`, `telnyx_call_control_application`, `telnyx_credential_connection`, `telnyx_custom_storage_credential`, `telnyx_dialogflow_connection`, `telnyx_document`, `telnyx_dynamic_emergency_address`, `telnyx_dynamic_emergency_endpoint`, `telnyx_external_connection`, `telnyx_external_connection_phone_number_settings`, `telnyx_fax_application`, `telnyx_fqdn_connection`, `telnyx_fqdn_connection_authentication`, `telnyx_inbound_voice_channels`, `telnyx_inexplicit_number_order`, `telnyx_ip_connection`, `telnyx_mobile_push_credential`, `telnyx_mobile_voice_connection`, `telnyx_number_block_order`, ~~`telnyx_number_order`~~, `telnyx_number_reservation`, ~~`telnyx_outbound_voice_profile`~~, `telnyx_pay_connector`, `telnyx_phone_number`, `telnyx_phone_number_messaging_settings`, ~~`telnyx_phone_number_voice_settings`~~, `telnyx_phone_number_voicemail`, `telnyx_porting_loa_configuration`, `telnyx_porting_order`, `telnyx_porting_order_additional_document`, `telnyx_porting_order_associated_phone_number`, `telnyx_porting_order_phone_number_block`, `telnyx_porting_order_phone_number_extension`, `telnyx_queue`, `telnyx_requirement_group`, `telnyx_siprec_connector`, `telnyx_telephony_credential`, `telnyx_texml_application`, `telnyx_texml_queue`, `telnyx_uac_connection`, `telnyx_user_address`, `telnyx_verified_number`, `telnyx_verify_profile`, `telnyx_verify_profile_message_template`, `telnyx_voice_channel_zone`

### Wave 2: Messaging

26 resources, 52 data sources, 25 actions, 0 ephemeral resources.

`telnyx_10dlc_brand`, `telnyx_10dlc_campaign`, `telnyx_10dlc_phone_number_campaign`, `telnyx_10dlc_shared_campaign`, `telnyx_alphanumeric_sender_id`, `telnyx_messaging_hosted_number`, `telnyx_messaging_hosted_number_order`, `telnyx_messaging_profile`, `telnyx_messaging_profile_auto_response`, `telnyx_messaging_rcs_agent`, `telnyx_messaging_tollfree_verification_request`, `telnyx_mobile_phone_number_settings`, `telnyx_notification_channel`, `telnyx_notification_profile`, `telnyx_notification_setting`, `telnyx_rcs_agent`, `telnyx_rcs_agent_test_device`, `telnyx_rcs_brand`, `telnyx_short_code_settings`, `telnyx_whatsapp_business_account_settings`, `telnyx_whatsapp_message_template`, `telnyx_whatsapp_phone_number_calling_routing`, `telnyx_whatsapp_phone_number_calling_settings`, `telnyx_whatsapp_phone_number_conversational_components`, `telnyx_whatsapp_phone_number_profile`, `telnyx_whatsapp_user_data`

### Wave 3: Networking and wireless

23 resources, 35 data sources, 15 actions, 2 ephemeral resources.

`telnyx_access_ip_address`, `telnyx_access_ip_range`, `telnyx_fqdn`, `telnyx_global_ip`, `telnyx_global_ip_assignment`, `telnyx_global_ip_health_check`, `telnyx_ip`, `telnyx_network`, `telnyx_network_default_gateway`, `telnyx_network_route`, `telnyx_private_wireless_gateway`, `telnyx_public_internet_gateway`, `telnyx_sim_card`, `telnyx_sim_card_data_usage_notification`, `telnyx_sim_card_group`, `telnyx_sim_card_order`, `telnyx_tailscale_subnet_router`, `telnyx_traffic_policy_profile`, `telnyx_tunnel`, `telnyx_virtual_cross_connect`, `telnyx_wireguard_interface`, `telnyx_wireguard_peer`, `telnyx_wireless_blocklist`

### Wave 4: Email

11 resources, 12 data sources, 7 actions, 0 ephemeral resources.

`telnyx_email_dedicated_ip`, `telnyx_email_domain`, `telnyx_email_domain_webhook`, `telnyx_email_inbox`, `telnyx_email_inbox_filter`, `telnyx_email_ip_pool`, `telnyx_email_ip_pool_ip`, `telnyx_email_suppression`, `telnyx_email_template`, `telnyx_email_unsubscribe_group`, `telnyx_email_unsubscribe_group_suppression`

### Wave 5: AI, storage and compute

40 resources, 49 data sources, 15 actions, 0 ephemeral resources.

~~`telnyx_ai_assistant`~~, `telnyx_ai_assistant_canary_deploy`, `telnyx_ai_assistant_tag`, `telnyx_ai_assistant_test`, `telnyx_ai_assistant_tool_attachment`, ~~`telnyx_ai_avatar`~~, `telnyx_ai_collection`, `telnyx_ai_collection_settings`, `telnyx_ai_collection_source`, `telnyx_ai_conversation_insight`, `telnyx_ai_conversation_insight_group`, `telnyx_ai_conversation_insight_group_assignment`, `telnyx_ai_gateway_token_group`, `telnyx_ai_gateway_token_key`, ~~`telnyx_ai_mcp_server`~~, `telnyx_ai_memory_namespace`, `telnyx_ai_memory_namespace_settings`, `telnyx_ai_mission`, `telnyx_ai_mission_knowledge_base`, `telnyx_ai_mission_mcp_server`, `telnyx_ai_mission_tool`, `telnyx_ai_test_agent`, `telnyx_ai_test_metric`, `telnyx_ai_test_scenario`, `telnyx_ai_test_schedule`, ~~`telnyx_ai_tool`~~, `telnyx_compute_function_log_export`, `telnyx_integration_secret`, `telnyx_media`, `telnyx_pronunciation_dictionary`, `telnyx_room`, `telnyx_storage_bucket_ssl_certificate`, `telnyx_storage_cloudfs_filesystem`, `telnyx_storage_kv_key`, `telnyx_storage_kv_namespace`, `telnyx_storage_migration`, `telnyx_storage_migration_source`, `telnyx_storage_sql_database`, `telnyx_voice_clone`, `telnyx_voice_design`

### Wave 6: Account and identity

9 resources, 81 data sources, 43 actions, 0 ephemeral resources.

`telnyx_authentication_provider`, `telnyx_auto_recharge_preferences`, `telnyx_display_identity_record`, `telnyx_enterprise`, `telnyx_managed_account`, `telnyx_oauth_client`, `telnyx_reputation_enterprise`, `telnyx_reputation_number`, `telnyx_spend_limit`

## Open questions and decisions

- **Releasing purchased assets on destroy (decided 2026-10-29 by Roman).** `telnyx_phone_number`, `telnyx_sim_card` and `telnyx_messaging_hosted_number` adopt objects that were bought or hosted elsewhere. Their `DELETE` releases the number or SIM, and that cannot be undone. By default, destroy only removes the object from state and warns. An optional `release_on_destroy = true` (default `false`) opts in to the real delete. Wave 1 specifies this in the `numbers` spec when it ships `telnyx_phone_number`.
- **Storage buckets.** The pinned spec has bucket SSL certificates, usage and presigned URLs, but no operation that creates a bucket. Buckets are created through Telnyx's S3-compatible API. Covering them means either a hand-written `telnyx_storage_bucket` over the S3 API, which is outside the OpenAPI contract, or documenting the AWS provider with a Telnyx endpoint. This is decided in wave 5.
