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

# Start the warm-up once the IP is in the pool. Needs Terraform 1.14 or later.
resource "terraform_data" "warmup" {
  input = telnyx_email_ip_pool_ip.a.dedicated_ip_id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.telnyx_email_dedicated_ip_start_warmup.a]
    }
  }
}

action "telnyx_email_dedicated_ip_start_warmup" "a" {
  config {
    id = telnyx_email_dedicated_ip.a.id
  }
}
