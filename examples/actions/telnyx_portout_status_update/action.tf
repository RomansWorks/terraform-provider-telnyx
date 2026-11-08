# Reject a port-out request. Needs Terraform 1.14 or later:
#   terraform apply -invoke=action.telnyx_portout_status_update.reject
action "telnyx_portout_status_update" "reject" {
  config {
    id     = var.portout_id
    status = "rejected-pending"
    reason = "The account PIN does not match."
  }
}

variable "portout_id" {
  type = string
}
