resource "telnyx_email_domain" "mail" {
  domain          = "mail.example.com"
  inbound_enabled = true

  tracking = {
    open_tracking        = false
    click_tracking       = false
    unsubscribe_tracking = true
  }

  # A verified domain can only be deleted with force_destroy = true.
  force_destroy = true
}

# The DNS records to publish at your DNS provider.
data "telnyx_email_domain_dns_records" "mail" {
  domain_id = telnyx_email_domain.mail.id
}
