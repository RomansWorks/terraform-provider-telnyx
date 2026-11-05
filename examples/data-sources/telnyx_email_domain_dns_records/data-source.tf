data "telnyx_email_domain_dns_records" "mail" {
  domain_id = telnyx_email_domain.mail.id
}
