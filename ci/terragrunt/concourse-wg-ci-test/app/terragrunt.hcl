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

  concourse_helm_version = local.config.concourse_helm_version

  gke_name = local.config.gke_name
  gke_workers_pool_machine_type = local.config.gke_workers_pool_machine_type
  gke_workers_pool_node_count = local.config.gke_workers_pool_node_count
  gke_default_pool_node_count = local.config.gke_default_pool_node_count

  load_balancer_ip = dependency.infra.outputs.load_balancer_ip
  load_balancer_dns = dependency.infra.outputs.load_balancer_dns

  concourse_github_mainTeam = local.config.concourse_github_mainTeam
  concourse_github_mainTeamUser = local.config.concourse_github_mainTeamUser

}