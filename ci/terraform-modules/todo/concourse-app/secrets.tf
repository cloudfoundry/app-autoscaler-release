# Provide github oauth token prior to app deployment
data "google_secret_manager_secret_version" "github_oauth" {
  secret = data.terraform_remote_state.infra.outputs.github_oauth.name
}

locals {
  github_oauth = yamldecode(data.google_secret_manager_secret_version.github_oauth.secret_data)
}

resource "kubernetes_secret_v1" "github_oauth" {
  metadata {
    name      = "github"
    namespace = var.concourse_app.namespace
  }
  data = {
    id     = local.github_oauth["id"]
    secret = local.github_oauth["secret"]
  }
}


