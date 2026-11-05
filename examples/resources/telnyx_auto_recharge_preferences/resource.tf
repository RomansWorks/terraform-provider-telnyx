# Top the balance up by 100 whenever it falls below 20. There is one per
# account.
resource "telnyx_auto_recharge_preferences" "this" {
  enabled          = true
  threshold_amount = "20.00"
  recharge_amount  = "100.00"
}
