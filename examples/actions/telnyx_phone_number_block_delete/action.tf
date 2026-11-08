# Release every number of a block. The job runs in the background: follow
# it with the telnyx_phone_number_block_jobs data source.
action "telnyx_phone_number_block_delete" "old_block" {
  config {
    phone_number_block_id = "f3946371-7199-4261-9c3d-81a0d7935146"
  }
}
