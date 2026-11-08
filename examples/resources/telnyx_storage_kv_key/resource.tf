resource "telnyx_storage_kv_namespace" "example" {
  name = "feature-flags"
}

resource "telnyx_storage_kv_key" "example" {
  namespace_id = telnyx_storage_kv_namespace.example.id
  key          = "new-checkout"
  value        = jsonencode({ enabled = true, rollout = 0.1 })
}
