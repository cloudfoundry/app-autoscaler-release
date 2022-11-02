# Provide github oauth token prior to app deployment
data "google_secret_manager_secret_version" "github_oauth" {
  secret = "${var.gke_name}-concourse-github-oauth"
  project = var.project
}

locals {
  github_oauth = yamldecode(data.google_secret_manager_secret_version.github_oauth.secret_data)
}

resource "kubernetes_secret_v1" "github_oauth" {
  metadata {
    name      = "github"
    namespace = "concourse"
  }
  data = {
    id     = local.github_oauth["id"]
    secret = local.github_oauth["secret"]
  }
}


