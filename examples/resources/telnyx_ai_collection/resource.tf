resource "telnyx_ai_collection" "example" {
  name        = "Support articles"
  description = "Help centre articles for the support assistant"
  retrieval = {
    top_k          = 5
    retrieval_type = "hybrid"
  }
}
