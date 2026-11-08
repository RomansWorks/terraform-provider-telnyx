resource "telnyx_ai_test_schedule" "example" {
  name               = "Nightly regression"
  crontab_expression = "0 2 * * *"
  timezone           = "UTC"
  agent              = 42
  scenarios          = [101, 102]
}
