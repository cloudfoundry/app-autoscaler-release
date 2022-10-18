data "carvel_ytt" "concourse_backend" {

  files = [
    "../../config/carvel-secretgen-controller",
    "../../config/database",
    "../../config/values",
  ]

  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
  }
}


resource "carvel_kapp" "concourse_backend" {
  app          = "concourse-backend"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.concourse_backend.result
  diff_changes = true

  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }

  delete {
    # WARN: if you change delete options you have to run terraform apply first.
    raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"SQLUser\",\"Namespace\"]}}}]}"]
  }
}
