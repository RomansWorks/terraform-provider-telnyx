# Telnyx cannot delete avatars: destroying this resource only removes it
# from Terraform state. A new image creates a new avatar.
resource "telnyx_ai_avatar" "portrait" {
  name       = "Front desk portrait"
  image_path = "${path.module}/portrait.png"
}
