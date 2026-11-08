# Check the DNS records again after publishing them. Needs Terraform 1.14 or later.
action "telnyx_email_domain_verify" "mail" {
  config {
    domain_id = telnyx_email_domain.mail.id
  }
}
