resource "telnyx_global_ip_health_check" "example" {
  global_ip_id      = telnyx_global_ip.example.id
  health_check_type = "http_status_2xx"
  health_check_params = jsonencode({
    path = "/health"
    port = 8080
  })
}
