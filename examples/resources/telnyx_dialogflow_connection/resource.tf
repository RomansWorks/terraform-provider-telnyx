resource "telnyx_dialogflow_connection" "example" {
  connection_id  = telnyx_call_control_application.ivr.id
  dialogflow_api = "cx"
  location       = "global"
  environment    = "production"

  service_account_wo         = var.dialogflow_service_account_json
  service_account_wo_version = 1
}

# The Google Cloud service account key file for the Dialogflow agent.
variable "dialogflow_service_account_json" {
  type      = string
  sensitive = true
}
