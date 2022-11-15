data "google_sql_database_instance" "concourse" {
  name    = var.sql_instance_name
  project = var.project

}

resource "google_sql_database" "concourse" {

  for_each = toset([
    "concourse",
    "credhub",
    "uaa"
  ])
  charset    = "UTF8"
  collation  = "en_US.UTF8"
  instance   = data.google_sql_database_instance.concourse.name
  name       = each.key
  project    = var.project
  depends_on = [data.google_sql_database_instance.concourse, carvel_kapp.sqlproxy, carvel_kapp.carvel_secretgen]

}

