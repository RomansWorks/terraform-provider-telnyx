data "telnyx_email_domain_health" "mail" {
  id = telnyx_email_domain.mail.id
}

output "mail_can_send" {
  value = data.telnyx_email_domain_health.mail.usable_for_sending
}
