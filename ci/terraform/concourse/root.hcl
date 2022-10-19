locals {
  #config = yamldecode(file(find_in_parent_folders("config.yaml")))
  tgconf = yamldecode(file("config.yaml"))
}

# remote_state {
#   backend = "gcs"
#   generate = {
#     path      = "backend.tf"
#     if_exists = "overwrite"
#   }
#   config = {
#     bucket         = "terraform-state-${local.tgconf.gke.name}"
#     prefix         = "terraform/state/root"
#     project        = "${local.tgconf.project}"
#     location       = "${local.tgconf.region}"
#     # use for uniform bucket-level access 
#     # (https://cloud.google.com/storage/docs/uniform-bucket-level-access)
#     enable_bucket_policy_only = false 
#   }
# }

