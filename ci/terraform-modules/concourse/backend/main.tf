data "carvel_ytt" "concourse_sqlproxy" {

  files = [ 
    "files/config/carvel-secretgen-controller",
    "files/config/cloud_sql",
   ]
  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
    "database.instance" = var.sql_instance_name
    "sql_proxy_account.name" = "${var.gke_name}-sql-proxy"
    "sql_proxy_account.email" = "${var.gke_name}-sql-proxy@${var.project}.iam.gserviceaccount.com"
  }
}


resource "carvel_kapp" "concourse_sqlproxy" {
  app          = "concourse-sqlproxy"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.concourse_sqlproxy.result
  diff_changes = true

  # use in maintenance only when needed (should not be required normally)
  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }

  # delete {
  #   # WARN: if you change delete options you have to rerun terraform apply first.
  #   raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
  # }

  depends_on = [ google_sql_database.concourse ]
}

#-----------------------------------------------------------------------------------------------------------------

data "carvel_ytt" "concourse_backend" {

  files = [ 
    "files/config/credhub",   
    "files/config/uaa"
   ]
  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
    "database.instance" = var.sql_instance_name
    "sql_proxy_account.name" = "${var.gke_name}-sql-proxy"
    "sql_proxy_account.email" = "${var.gke_name}-sql-proxy@${var.project}.iam.gserviceaccount.com"
  }
}


resource "carvel_kapp" "concourse_backend" {
  app          = "concourse-backend"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.concourse_backend.result
  diff_changes = true

  # use in maintenance only when needed (should not be required normally)
  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }

  # delete {
  #   # WARN: if you change delete options you have to rerun terraform apply first.
  #   raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
  # }

  depends_on = [ carvel_kapp.concourse_sqlproxy ]
}