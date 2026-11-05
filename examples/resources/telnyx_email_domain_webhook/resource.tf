resource "telnyx_email_domain_webhook" "events" {
  domain_id = telnyx_email_domain.mail.id
  url       = "https://hooks.example.com/telnyx/email"
  events    = ["email.delivered", "email.bounced"]
}
