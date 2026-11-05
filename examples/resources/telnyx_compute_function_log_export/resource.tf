resource "telnyx_compute_function_log_export" "example" {
  function_id               = "fn-123"
  endpoint                  = "https://logs.example.com/ingest"
  headers                   = jsonencode({ Authorization = "Bearer ${var.log_token}" })
  runtime_export_enabled    = true
  invocation_export_enabled = true
}

variable "log_token" {
  type      = string
  sensitive = true
}
