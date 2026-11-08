# A dedicated IP is billed monthly from purchase until destroy releases it.
resource "telnyx_email_dedicated_ip" "a" {
  region = "ewr-nj1"
}

resource "telnyx_email_ip_pool" "marketing" {
  name      = "marketing"
  domain_id = telnyx_email_domain.mail.id
}

resource "telnyx_email_ip_pool_ip" "a" {
  pool_id         = telnyx_email_ip_pool.marketing.id
  dedicated_ip_id = telnyx_email_dedicated_ip.a.id
}

# To start the warm-up, see the telnyx_email_dedicated_ip_start_warmup action.
