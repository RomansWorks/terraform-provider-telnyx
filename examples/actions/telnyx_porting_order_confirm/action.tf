# Submit the porting order once it is created. Needs Terraform 1.14 or later.
resource "terraform_data" "submit" {
  input = telnyx_porting_order.office.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.telnyx_porting_order_confirm.office]
    }
  }
}

action "telnyx_porting_order_confirm" "office" {
  config {
    id = telnyx_porting_order.office.id
  }
}
