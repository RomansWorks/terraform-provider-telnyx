---
page_title: "Storage buckets"
subcategory: ""
description: |-
  Create Telnyx Storage buckets with the AWS provider and manage their SSL certificates with this provider.
---

# Storage buckets

Telnyx Storage buckets are created through Telnyx's S3-compatible API, not the Telnyx API v2 this provider is built on. Manage the buckets themselves with the [AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) pointed at Telnyx's S3 endpoint, and the rest of Telnyx Storage with this provider: `telnyx_storage_bucket_ssl_certificate`, `telnyx_storage_migration` and its sources, KV namespaces, SQL databases and CloudFS filesystems.

## Configure the AWS provider for Telnyx

Telnyx Storage takes your Telnyx API key as the S3 access key ID and ignores the secret key and the request signature, so passing the API key as both is the simplest setup. Each bucket is homed in one region, and every call except listing buckets must go to that region's endpoint, `https://<region>.telnyxcloudstorage.com`. The regions are `us-central-1`, `us-east-1`, `us-west-1`, `eu-central-1`, `ap-southeast-1` and `ca-central-1`; see the [Telnyx Storage quick start](https://developers.telnyx.com/docs/cloud-storage/quick-start) and [API endpoints](https://developers.telnyx.com/docs/cloud-storage/api-endpoints) for the current list. Use one aliased AWS provider per region.

```terraform
provider "aws" {
  alias  = "telnyx"
  region = "us-central-1"

  access_key = var.telnyx_api_key
  secret_key = var.telnyx_api_key

  endpoints {
    s3 = "https://us-central-1.telnyxcloudstorage.com"
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true
}

variable "telnyx_api_key" {
  type      = string
  sensitive = true
}
```

## Create a bucket and give it a certificate

```terraform
resource "aws_s3_bucket" "assets" {
  provider = aws.telnyx
  bucket   = "example-assets"
}

resource "telnyx_storage_bucket_ssl_certificate" "assets" {
  bucket_name     = aws_s3_bucket.assets.bucket
  certificate_pem = file("${path.module}/cdn.example.com.crt")
  private_key_pem = file("${path.module}/cdn.example.com.key")
}
```

SSL certificates, public buckets and presigned URLs are available only in the US, APAC and CA regions, so `telnyx_storage_bucket_ssl_certificate` fails for a bucket in `eu-central-1`.

Only the S3 features Telnyx Storage implements work through the AWS provider, and support differs by region; check the [compatibility matrix](https://developers.telnyx.com/docs/cloud-storage/supported). Resources such as bucket policies or lifecycle rules fail when applied where Telnyx does not support them. Telnyx does not document whether path-style addressing is required; if bucket requests fail to resolve, set `s3_use_path_style = true` in the provider block.
