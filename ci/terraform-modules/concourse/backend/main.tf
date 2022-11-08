data "carvel_ytt" "sqlproxy_secretgen" {

  files = [
    "files/config/carvel-secretgen-controller",
    "files/config/cloud_sql_proxy",
  ]
  values = {
    "google.project_id"       = var.project
    "google.region"           = var.region
    "database.instance"       = var.sql_instance_name
    "sql_proxy_account.name"  = "${var.gke_name}-sql-proxy"
  }
}


resource "carvel_kapp" "sqlproxy_secretgen" {
  app          = "sqlproxy-secretgen"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.sqlproxy_secretgen.result
  diff_changes = true

  # use in maintenance only when needed (should not be required normally)
  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }
  # delete {
  #   # WARN: if you change delete options you have to rerun terraform apply first.
  #   raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
  # }
}

#-----------------------------------------------------------------------------------------------------------------

data "carvel_ytt" "credhub_uaa" {

  files = [
    "files/config/credhub",
    "files/config/uaa"
  ]
  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
  }
}


resource "carvel_kapp" "credhub_uaa" {
  app          = "credhub-uaa"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.credhub_uaa.result
  diff_changes = true
  depends_on   = [google_service_account.sql_proxy, carvel_kapp.sqlproxy_secretgen, google_sql_database.concourse]

  # use in maintenance only when needed (should not be required normally)
  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }
  # delete {
  #   # WARN: if you change delete options you have to rerun terraform apply first.
  #   raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
  # }
}