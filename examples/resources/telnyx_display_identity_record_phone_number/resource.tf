resource "telnyx_display_identity_record_phone_number" "main" {
  dir_id       = telnyx_display_identity_record.acme.id
  phone_number = "+13125550100"

  documents = [{
    document_id   = telnyx_document.loa.id
    document_type = "letter_of_authorization"
  }]
}
