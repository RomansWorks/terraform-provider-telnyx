resource "telnyx_storage_migration_source" "example" {
  source_provider = "aws"
  source_region   = "us-east-1"
  bucket_name     = "legacy-assets"
  provider_auth = {
    access_key        = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}

resource "telnyx_storage_migration" "example" {
  source_id          = telnyx_storage_migration_source.example.id
  target_bucket_name = "example-assets"
  target_region      = "us-central-1"
}

variable "aws_access_key_id" {
  type = string
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
