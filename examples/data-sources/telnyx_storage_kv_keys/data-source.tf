data "telnyx_storage_kv_keys" "flags" {
  namespace_id = telnyx_storage_kv_namespace.flags.id
  prefix       = "feature-"
}
