resource "telnyx_call_control_application" "ivr" {
  application_name  = "ivr"
  webhook_event_url = "https://example.com/telnyx/webhooks"
}
