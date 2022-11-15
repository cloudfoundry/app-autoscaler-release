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
    prefix         = "${local.config.gcs_prefix}/concourse-infra"
    project        = "${local.config.project}"
    location       = "${local.config.region}"
    # use for uniform bucket-level access
    # (https://cloud.google.com/storage/docs/uniform-bucket-level-access)
    enable_bucket_policy_only = true
  }
}

# git for teams
terraform {
  source = local.config.tf_modules.infra
}

inputs = {
  project = local.config.project
  region  = local.config.region
  zone    = local.config.zone

  gke_name = local.config.gke_name
  gke_controlplane_version = local.config.gke_controlplane_version
  gke_cluster_ipv4_cidr = local.config.gke_cluster_ipv4_cidr
  gke_services_ipv4_cidr_block = local.config.gke_services_ipv4_cidr_block
  gke_master_ipv4_cidr_block = local.config.gke_master_ipv4_cidr_block
  gke_subnet_cidr = local.config.gke_subnet_cidr

  gke_default_pool_machine_type = local.config.gke_default_pool_machine_type
  gke_default_pool_ssd_count = local.config.gke_default_pool_ssd_count
  gke_default_pool_node_count = local.config.gke_default_pool_node_count
  gke_default_pool_autoscaling_max = local.config.gke_default_pool_autoscaling_max

  gke_workers_pool_machine_type = local.config.gke_workers_pool_machine_type
  gke_workers_pool_ssd_count = local.config.gke_workers_pool_ssd_count
  gke_workers_pool_node_count = local.config.gke_workers_pool_node_count
  gke_workers_pool_autoscaling_max = local.config.gke_workers_pool_autoscaling_max

  gke_cloud_nat_min_ports_per_vm = local.config.gke_cloud_nat_min_ports_per_vm

  sql_instance_name = "${local.config.gke_name}-concourse"
  sql_instance_tier = local.config.sql_instance_tier
  sql_instance_disk_size = local.config.sql_instance_disk_size
  sql_instance_backup_location = local.config.sql_instance_backup_location
  sql_instance_secondary_zone = local.config.secondary_zone

  dns_record = local.config.dns_record
  dns_zone = local.config.dns_zone
  dns_domain = local.config.dns_domain
  dns_name  = "${local.config.dns_zone}-${local.config.dns_domain}."

  wg_ci_human_account_permissions = local.config.wg_ci_human_account_permissions
  
  github_secret_name = "${local.config.gke_name}-concourse-github-oauth"
}