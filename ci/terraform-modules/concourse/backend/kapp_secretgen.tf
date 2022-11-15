data "carvel_ytt" "carvel_secretgen" {

  files = [
    "files/config/carvel-secretgen-controller",
  ]
  values = {
    "google.project_id"       = var.project
    "google.region"           = var.region
  }
}


resource "carvel_kapp" "carvel_secretgen" {
  app          = "carvel-secretgen"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.carvel_secretgen.result
  diff_changes = true
  depends_on = [kubernetes_namespace.concourse]

  deploy {
    raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  }
}

