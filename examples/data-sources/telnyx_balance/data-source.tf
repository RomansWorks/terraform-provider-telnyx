data "telnyx_balance" "account" {}

output "available_credit" {
  value = data.telnyx_balance.account.available_credit
}
