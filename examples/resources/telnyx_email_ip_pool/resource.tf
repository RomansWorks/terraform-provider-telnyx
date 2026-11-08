resource "telnyx_email_ip_pool" "default" {
  name       = "default"
  is_default = true

  # Deleting a pool that still has IPs needs force_destroy = true.
  force_destroy = true
}
