## ADDED Requirements

### Requirement: Reserved filter names
A list data source's query parameter whose snake_case name is reserved by Terraform (a meta-argument such as `provider` or `count`) or by the list itself (`max_items`) SHALL become the filter attribute `<name>_filter`, sent under the parameter's own name. A path parameter with a reserved name, or a renamed filter that collides with another parameter, SHALL be a generation error.

#### Scenario: Provider filter
- **WHEN** `data "telnyx_speech_to_text_providers" "x" { provider_filter = "deepgram" }` is read
- **THEN** the request is `GET /speech-to-text/providers?provider=deepgram`

#### Scenario: Collision
- **WHEN** an operation takes both `provider` and `provider_filter` as query parameters
- **THEN** generation fails naming both parameters
