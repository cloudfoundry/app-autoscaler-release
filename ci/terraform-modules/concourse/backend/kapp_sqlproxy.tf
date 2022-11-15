data "carvel_ytt" "sqlproxy" {

  files = [
    "files/config/cloud_sql_proxy",
  ]
  values = {
    "google.project_id"       = var.project
    "google.region"           = var.region
    "database.instance"       = var.sql_instance_name
    "sql_proxy_account.name"  = "${var.gke_name}-sql-proxy"
  }
}


resource "carvel_kapp" "sqlproxy" {
  app          = "sqlproxy"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.sqlproxy.result
  diff_changes = true

  depends_on = [ kubectl_manifest.config_connector, carvel_kapp.carvel_secretgen, kubernetes_namespace.concourse ]

  deploy {
    raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  }
}

#-----------------------------------------------------------------------------------------------------------------

