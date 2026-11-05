## ADDED Requirements

### Requirement: Function log exports
The provider SHALL manage `telnyx_compute_function_log_export`, the log export of a function identified and imported by `function_id`. Create and update SHALL use `PUT /compute/funcs/{id}/logs/export`. `headers` SHALL be sensitive and, since the API does not return them, kept from configuration. The provider SHALL also offer `data "telnyx_compute_functions"` and `data "telnyx_compute_function_revisions"`.

#### Scenario: Change endpoint
- **WHEN** `endpoint` changes
- **THEN** the provider sends the PUT again and the next plan is empty

#### Scenario: Destroy
- **WHEN** the export is destroyed
- **THEN** `DELETE /compute/funcs/{id}/logs/export` is sent
