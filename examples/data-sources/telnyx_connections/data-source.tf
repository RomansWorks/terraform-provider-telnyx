# Connections of every type whose name contains "office".
data "telnyx_connections" "office" {
  filter_connection_name_contains = "office"
}
