resource "telnyx_email_suppression" "complainer" {
  to        = "someone@example.org"
  domain_id = telnyx_email_domain.mail.id
}
