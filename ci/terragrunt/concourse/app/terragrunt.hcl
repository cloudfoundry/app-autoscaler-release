dependencies {
  paths = ["../infra", "../backend"]
}

dependency "infra" {
  config_path = "../infra"
}

locals {
  config = yamldecode(file("../config.yaml"))
}

remote_state {
  backend = "gcs"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite"
  }
  config = {
    bucket         = "${local.config.gcs_bucket}"
    prefix         = "${local.config.gcs_prefix}/concourse-app"
    project        = "${local.config.project}"
    location       = "${local.config.region}"
    # use for uniform bucket-level access
    # (https://cloud.google.com/storage/docs/uniform-bucket-level-access)
    enable_bucket_policy_only = false
  }
}

# git for teams
terraform {
  source = local.config.tf_modules.app
}

inputs = {
  project = local.config.project
  region  = local.config.region
  zone    = local.config.zone

  gke_name = local.config.gke_name

  github_oauth_secret_name = dependency.infra.outputs.github_oauth_secret_name

}