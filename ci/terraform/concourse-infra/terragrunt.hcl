locals {
  tgconf = yamldecode(file("../config.yaml"))
}

# include "root" {
#   path = find_in_parent_folders()
# }

remote_state {
  backend = "gcs"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite"
  }
  config = {
    bucket         = "terraform-state-${local.tgconf.gke.name}"
    prefix         = "concourse-infra"
    project        = "${local.tgconf.project}"
    location       = "${local.tgconf.region}"
    # use for uniform bucket-level access 
    # (https://cloud.google.com/storage/docs/uniform-bucket-level-access)
    enable_bucket_policy_only = false 
  }
}

# terraform {
#   source = "git::${local.tgconf.source.concourse_infra.url}?ref=${local.tgconf.source.concourse_infra.ref}"
# }

inputs = {
  project = local.tgconf.project
  #tgconf = yamldecode(file("../config.yaml"))
  #module_sources = local.tgconf.module_sources
 }