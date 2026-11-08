# The status of a bulk update or release job, and the numbers that failed.
data "telnyx_phone_number_job" "update" {
  id = "42587e44-3a3e-46de-9255-0c9a7a1d1ec7"
}

output "failed_numbers" {
  value = data.telnyx_phone_number_job.update.failed_operations[*].phone_number
}
