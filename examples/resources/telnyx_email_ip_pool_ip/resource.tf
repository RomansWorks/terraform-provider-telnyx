resource "telnyx_email_ip_pool_ip" "a" {
  pool_id         = telnyx_email_ip_pool.default.id
  dedicated_ip_id = telnyx_email_dedicated_ip.a.id
  weight          = 100
}
