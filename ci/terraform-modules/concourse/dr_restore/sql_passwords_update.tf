data "kubernetes_secret_v1" "sql_user_password" {

    for_each = toset([
    "concourse",
    "credhub",
    "uaa"
  ])

  metadata {
    name      = "${each.key}-postgresql-password"
    namespace = "concourse"
  }

  binary_data = {
    "password" = ""
  }
}

resource "google_sql_user" "sql_user_pass_restored" {
  instance = var.sql_instance_name
  project  = var.project

  for_each = toset([
    "concourse",
    "credhub",
    "uaa"
  ])

  # in this case we have the same names for users and dbs
  name     = each.key
  password = base64decode(data.kubernetes_secret_v1.sql_user_password[each.key].binary_data.password)

}