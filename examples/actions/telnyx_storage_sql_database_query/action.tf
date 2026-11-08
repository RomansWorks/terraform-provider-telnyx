# Create the schema once the database exists. Needs Terraform 1.14 or later.
resource "terraform_data" "schema" {
  input = telnyx_storage_sql_database.orders.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.telnyx_storage_sql_database_query.schema]
    }
  }
}

action "telnyx_storage_sql_database_query" "schema" {
  config {
    id  = telnyx_storage_sql_database.orders.id
    sql = "CREATE TABLE IF NOT EXISTS orders (id TEXT PRIMARY KEY, total REAL)"
  }
}
