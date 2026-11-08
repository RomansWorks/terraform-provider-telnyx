resource "telnyx_storage_migration_source" "legacy" {
  source_provider = "aws"
  source_region   = "us-east-1"
  bucket_name     = "legacy-assets"
  provider_auth = {
    access_key        = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}

variable "aws_access_key_id" {
  type = string
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
