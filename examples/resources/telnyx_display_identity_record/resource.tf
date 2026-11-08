# The name and call reasons shown when Acme calls its customers.
resource "telnyx_display_identity_record" "acme" {
  enterprise_id             = telnyx_enterprise.acme.id
  display_name              = "Acme"
  call_reasons              = ["Appointment reminders", "Delivery updates"]
  authorizer_name           = "Ada Lovelace"
  authorizer_email          = "ada@acme.example.com"
  certify_brand_is_accurate = true
  certify_no_shaft_content  = true
  certify_ip_ownership      = true
}
