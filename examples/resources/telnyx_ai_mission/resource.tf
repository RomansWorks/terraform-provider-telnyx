resource "telnyx_ai_mission" "example" {
  name           = "Weekly churn review"
  instructions   = "Review last week's cancelled accounts and summarise the reasons."
  execution_mode = "managed"
}
