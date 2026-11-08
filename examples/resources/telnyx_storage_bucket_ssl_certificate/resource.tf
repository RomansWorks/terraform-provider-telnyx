# Telnyx Storage buckets are created through its S3-compatible API, for
# example with the AWS provider (see the Storage buckets guide).
resource "aws_s3_bucket" "assets" {
  bucket = "example-assets"
}

resource "telnyx_storage_bucket_ssl_certificate" "example" {
  bucket_name     = aws_s3_bucket.assets.bucket
  certificate_pem = file("${path.module}/cdn.example.com.crt")
  private_key_pem = file("${path.module}/cdn.example.com.key")
}
