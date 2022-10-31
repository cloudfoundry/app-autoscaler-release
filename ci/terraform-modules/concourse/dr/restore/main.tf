# Ensure concourse-backend has been provisioned prior to running restore
#   otherwise it will complain about missing sql user secrets


# --- Credhub encryption key
data "google_secret_manager_secret_version" "credhub_encryption_key" {
  project = var.project
  secret = "${var.gke_name}-credhub-encryption-key"
}

resource "kubernetes_secret_v1" "credhub_encryption_key" {
  metadata {
    name      = "credhub-encryption-key"
    namespace = "concourse"
  }

  data = {
    password = data.google_secret_manager_secret_version.credhub_encryption_key.secret_data
  }
}

# --- SQL user passwords

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

