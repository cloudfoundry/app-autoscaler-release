data "helm_template" "concourse" {
  name   = "concourse"
  chart  = "../../build/concourse/_vendir"
  values = ["${file("../../build/concourse/values.yml")}"]

  set {
    name  = "concourse.web.externalUrl"
    value = var.dns_address.url
  }

  set {
    name  = "web.service.api.loadBalancerIP"
    value = data.terraform_remote_state.infra.outputs.load_balancer_ip
  }

  set {
    name  = "concourse.web.auth.mainTeam.github.team"
    value = var.concourse_app.github_mainTeam
  }

  set {
    name  = "concourse.web.auth.mainTeam.github.user"
    value = var.concourse_app.github_mainTeamUser
  }

  set {
    # For security reasons, remove any local users
    name  = "concourse.web.auth.mainTeam.localUser"
    value = ""
  }
}

data "carvel_ytt" "concourse_helm_ytt" {
  files                   = ["../../build/concourse/scrub_default_creds.yml"]
  config_yaml             = data.helm_template.concourse.manifest
  ignore_unknown_comments = true
}

resource "local_file" "concourse_rendered" {
  content         = data.carvel_ytt.concourse_helm_ytt.result
  filename        = "./config/concourse/_ytt_lib/concourse/rendered.yml"
  file_permission = "0644"
}


data "carvel_ytt" "concourse_app" {

  files = [
    "./config/concourse",
    "../../config/concourse/secrets",
    "../../config/concourse/concourse.yml",
    "../../config/credhub",
    "../../config/uaa",
    "../../config/values",
  ]

  values = {
    "google.project_id" = var.project
    "google.region"     = var.region
  }

  depends_on = [local_file.concourse_rendered]
}


resource "carvel_kapp" "concourse_app" {
  app          = var.concourse_app.kapp_app
  namespace    = var.concourse_app.namespace
  config_yaml  = data.carvel_ytt.concourse_app.result
  diff_changes = true

  # deploy {
  #   raw_options = ["--dangerous-override-ownership-of-existing-resources"]
  # }

  delete {
    # WARN: if you change delete options you have to run terraform apply first.
    raw_options = ["--filter={\"and\":[{\"not\":{\"resource\":{\"kinds\":[\"Namespace\"]}}}]}"]
  }

  depends_on = [kubernetes_secret_v1.github_oauth]
}

