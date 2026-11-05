## 1. Generator

- [x] 1.1 List filters with a reserved name take the suffix `_filter`; collisions and reserved path parameters stay errors; test

## 2. Wave 7 overrides

- [x] 2.1 Billing data sources: `telnyx_balance`, `telnyx_invoice(s)`, `telnyx_charges_summary`, `telnyx_charges_breakdown`, `telnyx_bundle_pricing_*`
- [x] 2.2 Billing actions: `telnyx_user_bundles_create`, `telnyx_user_bundle_deactivate`
- [x] 2.3 Reference data: `telnyx_regions`, `telnyx_country_coverage(_country)`, `telnyx_inventory_coverage`, `telnyx_pricing_product(s)`, `telnyx_requirement_type(s)`, `telnyx_requirement(s)`, `telnyx_mobile_network_operators`, `telnyx_noise_suppression_engines`, `telnyx_speech_to_text_providers`, `telnyx_call_reasons`
- [x] 2.4 Number operations data sources: `telnyx_available_phone_numbers`, `telnyx_available_phone_number_blocks`, `telnyx_number_order_phone_number(s)`, `telnyx_sub_number_order(s)`, `telnyx_regulatory_requirements`, `telnyx_portout(s)`, `telnyx_portout_rejections`, `telnyx_portout_report`, `telnyx_portout_supporting_documents`, `telnyx_infringement_claim`
- [x] 2.5 Number operations actions: `telnyx_number_order_phone_number_update`, `telnyx_sub_number_order_update`, `telnyx_sub_number_order_cancel`, `telnyx_phone_numbers_bulk_delete`, `telnyx_phone_numbers_bulk_update`, `telnyx_phone_numbers_bulk_update_emergency_settings`, `telnyx_portout_status_update`, `telnyx_portout_supporting_document_add`, `telnyx_requirement_version_schedule`, `telnyx_requirement_version_cancel`, `telnyx_infringement_claim_contest`
- [x] 2.6 Account lookups: `telnyx_connection(s)`, `telnyx_connection_count`, `telnyx_organization_user(s)`, `telnyx_terms_of_service_*`, `telnyx_user_tags`, `telnyx_ota_update(s)`, `telnyx_compute_actor`; actions `telnyx_organization_user_remove`, `telnyx_terms_of_service_branded_calling_agree`, `telnyx_terms_of_service_number_reputation_agree`
- [x] 2.7 AI integrations: `telnyx_ai_integration(s)`, `telnyx_ai_integration_connection(s)`, `telnyx_ai_integration_connection_delete`

## 3. Tests, docs and release notes

- [x] 3.1 `make generate` docs; examples for the main data sources and actions
- [x] 3.2 Changelog entry (`changie new`)
- [x] 3.3 `make check` passes (acceptance tests not run: no API key)
