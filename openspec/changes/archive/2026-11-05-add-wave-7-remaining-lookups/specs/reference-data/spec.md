## ADDED Requirements

### Requirement: Coverage and pricing
The provider SHALL offer read-only data sources for regions (`telnyx_regions`), number coverage (`telnyx_country_coverage`, `telnyx_country_coverage_country`, `telnyx_inventory_coverage`) and product prices (`telnyx_pricing_products`, `telnyx_pricing_product`). Lists SHALL page as other list data sources do.

#### Scenario: Country coverage
- **WHEN** `data "telnyx_country_coverage_country" "de" { country_code = "DE" }` is read
- **THEN** the provider sends `GET /country_coverage/countries/DE`

#### Scenario: Pricing of a product
- **WHEN** `data "telnyx_pricing_product" "p" { slug = "voice" }` is read
- **THEN** the provider pages through `GET /pricing/products/voice` and lists the entries

### Requirement: Catalogs
The provider SHALL offer read-only data sources for requirement types and requirements (`telnyx_requirement_types`, `telnyx_requirement_type`, `telnyx_requirements`, `telnyx_requirement`), mobile network operators, noise suppression engines, speech-to-text providers and call reasons. The speech-to-text provider filter SHALL be `provider_filter`.

#### Scenario: Speech-to-text providers
- **WHEN** `data "telnyx_speech_to_text_providers" "d" { provider_filter = "deepgram" }` is read
- **THEN** the provider sends `GET /speech-to-text/providers?provider=deepgram`
