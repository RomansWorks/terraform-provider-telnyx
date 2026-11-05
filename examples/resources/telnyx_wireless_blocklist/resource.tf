resource "telnyx_wireless_blocklist" "no_roaming" {
  name   = "no-roaming"
  type   = "country"
  values = ["CA", "MX"]
}
