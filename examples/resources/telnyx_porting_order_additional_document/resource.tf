resource "telnyx_document" "invoice" {
  url = "https://example.com/invoices/2026-10.pdf"
}

resource "telnyx_porting_order_additional_document" "invoice" {
  porting_order_id = telnyx_porting_order.office.id
  document_id      = telnyx_document.invoice.id
  document_type    = "invoice"
}
