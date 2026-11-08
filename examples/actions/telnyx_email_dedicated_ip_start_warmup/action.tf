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
