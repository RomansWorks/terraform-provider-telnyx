resource "telnyx_mobile_push_credential" "example" {
  type                      = "android"
  alias                     = "acme-android"
  project_account_json_file = var.firebase_service_account_json
}

# The Firebase service account key file, as downloaded from Google Cloud.
variable "firebase_service_account_json" {
  type      = string
  sensitive = true
}
