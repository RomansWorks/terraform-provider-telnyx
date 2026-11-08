# Assign a number to a regulatory bundle. Leave bundle_id unset to remove it
# from its bundle; a number cannot move between bundles in one step.
action "telnyx_phone_number_bundle_status_change" "assign" {
  config {
    id        = telnyx_phone_number.main.id
    bundle_id = "5194d8fc-87e6-4188-baa9-1c434bbe861b"
  }
}

action "telnyx_phone_number_bundle_status_change" "remove" {
  config {
    id = telnyx_phone_number.main.id
  }
}
