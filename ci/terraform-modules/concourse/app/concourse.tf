data "helm_template" "concourse" {
  name        = "concourse"
  repository  = "https://concourse-charts.storage.googleapis.com/"
  chart       = "concourse"
  version     = var.concourse_helm_version
  values      = ["${file("files/${var.gke_workers_pool_machine_type}.yml")}"]

  set {
    name  = "concourse.web.externalUrl"
    value = "https://${var.load_balancer_dns}"
  }

  set {
    name  = "web.service.api.loadBalancerIP"
    value = var.load_balancer_ip
  }

  set {
    name  = "concourse.web.auth.mainTeam.github.team"
    value = var.concourse_github_mainTeam
  }

  set {
    name  = "concourse.web.auth.mainTeam.github.user"
    value = var.concourse_github_mainTeamUser
  }

  set {
    # For security reasons, remove any local users
    name  = "concourse.web.auth.mainTeam.localUser"
    value = ""
  }

  set {
    name = "worker.replicas"
    value = var.gke_workers_pool_node_count
  }

  set {
    name = "web.replicas"
    value = var.gke_default_pool_node_count
  }
}

data "carvel_ytt" "concourse_app" {

  files = [ "files/config/concourse" ]

  config_yaml = data.helm_template.concourse.manifest

  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
  }
 }


resource "carvel_kapp" "concourse_app" {
  app          = "concourse-app"
  namespace    = "concourse"
  config_yaml  = data.carvel_ytt.concourse_app.result
  diff_changes = true

  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }

#   delete {
#     # WARN: if you change delete options you have to run terraform apply first.
#     raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
#   }

  depends_on = [kubernetes_secret_v1.github_oauth, carvel_kapp.credhub_uaa]
}

