# Email the operations team when the balance falls below 50. The API cannot
# change a setting, so every change replaces it.
data "telnyx_notification_event_conditions" "all" {}

locals {
  low_balance = one([
    for c in data.telnyx_notification_event_conditions.all.notification_event_conditions :
    c if c.name == "Low Balance"
  ])
}

resource "telnyx_notification_setting" "low_balance" {
  notification_profile_id         = telnyx_notification_profile.ops.id
  notification_channel_id         = telnyx_notification_channel.ops_email.id
  notification_event_condition_id = local.low_balance.id
  parameters = [
    { name = "amount", value = "50" },
  ]
}
