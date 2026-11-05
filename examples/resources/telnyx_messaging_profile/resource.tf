resource "telnyx_messaging_profile" "support" {
  name                     = "Support"
  whitelisted_destinations = ["US", "CA"]
}
