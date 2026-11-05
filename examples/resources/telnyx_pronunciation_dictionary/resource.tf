resource "telnyx_pronunciation_dictionary" "example" {
  name = "Brand names"
  items = [
    { type = "alias", text = "Telnyx", alias = "tell-nix" },
  ]
}
