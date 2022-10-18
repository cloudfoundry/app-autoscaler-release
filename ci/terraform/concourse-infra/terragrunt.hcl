locals {
  config = yamldecode(file("../config.yaml"))
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
    bucket         = "${local.config.gcs_bucket}"
    prefix         = "${local.config.gcs_prefix}/concourse-infra"
    project        = "${local.config.project}"
    location       = "${local.config.region}"
    # use for uniform bucket-level access
    # (https://cloud.google.com/storage/docs/uniform-bucket-level-access)
    enable_bucket_policy_only = false
  }
}

# terraform {
#   source = "git::${local.tgconf.source.concourse_infra.url}?ref=${local.tgconf.source.concourse_infra.ref}"
# }

 inputs = {
    project = local.config.project
    region  = local.config.region
    zone    = local.config.zone
    gke_name = local.config.gke_name

    sql_instance_name = "${local.config.gke_name}-concourse"
    sql_instance_tier = local.config.sql_instance_tier
    sql_instance_disk_size = local.config.sql_instance_disk_size
    sql_instance_backup_location = local.config.sql_instance_backup_location
    sql_instance_secondary_zone = local.config.secondary_zone

    vpc_name = local.config.vpc_name
    subnet_name = local.config.subnetwork_name

    dns_record = local.config.dns_record
    dns_zone = local.config.dns_zone
    dns_domain = local.config.dns_domain
    dns_name  = "${local.config.dns_zone}-${local.config.dns_domain}."

#   #tgconf = yamldecode(file("../config.yaml"))
#   #module_sources = local.tgconf.module_sources
  }


# terraform {
#     extra_arguments "custom_vars" {
#     commands = [
#       "apply",
#       "plan",
#       "import",
#       "refresh"
#     ]

#   arguments = [
#     "-var-file=${get_terragrunt_dir()}/../config.vars"
#   ]
#     }
#  }