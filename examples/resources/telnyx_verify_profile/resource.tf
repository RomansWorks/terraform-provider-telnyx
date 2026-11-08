resource "telnyx_verify_profile" "example" {
  name        = "Acme login"
  language    = "en-US"
  webhook_url = "https://example.com/webhooks/verify"

  sms = {
    app_name                          = "Acme"
    code_length                       = 6
    default_verification_timeout_secs = 300
    whitelisted_destinations          = ["US", "CA"]
  }
}
