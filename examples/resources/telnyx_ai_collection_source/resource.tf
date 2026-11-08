resource "telnyx_ai_collection_source" "example" {
  collection_id = telnyx_ai_collection.example.id
  source_type   = "bucket"
  bucket_id     = "policy-docs"
}
