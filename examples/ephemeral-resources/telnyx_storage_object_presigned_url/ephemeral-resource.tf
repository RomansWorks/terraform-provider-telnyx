# A link to one object that works without credentials for an hour. It is
# created on every run and never stored.
ephemeral "telnyx_storage_object_presigned_url" "report" {
  bucket_name = "reports"
  object_name = "2026-11.csv"
  ttl         = 3600
}
