# Store the connection's call recordings in your own S3 bucket.
resource "telnyx_custom_storage_credential" "example" {
  connection_id = telnyx_call_control_application.ivr.id
  backend       = "s3"
  configuration = jsonencode({
    backend               = "s3"
    bucket                = "acme-call-recordings"
    region                = "us-east-1"
    aws_access_key_id     = var.aws_access_key_id
    aws_secret_access_key = var.aws_secret_access_key
  })
}

variable "aws_access_key_id" {
  type = string
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
