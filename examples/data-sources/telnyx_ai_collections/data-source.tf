data "telnyx_ai_collections" "all" {}

output "collection_slugs" {
  value = [for c in data.telnyx_ai_collections.all.ai_collections : c.slug]
}
