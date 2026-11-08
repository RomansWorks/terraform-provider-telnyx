# Upload a document from a public URL, such as a letter of authorization for a
# porting order.
resource "telnyx_document" "example" {
  url      = "https://example.com/documents/loa.pdf"
  filename = "loa.pdf"
}
