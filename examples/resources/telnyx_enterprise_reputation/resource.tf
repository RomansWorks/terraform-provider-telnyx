# Monitor the reputation of Acme's numbers weekly, with a signed Letter of
# Authorization uploaded as a document.
resource "telnyx_enterprise_reputation" "acme" {
  enterprise_id   = telnyx_enterprise.acme.id
  loa_document_id = telnyx_document.reputation_loa.id
  check_frequency = "weekly"
}
