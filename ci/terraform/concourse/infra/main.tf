
module "concourse-infra" {
  #variables can't be used here
  #source = "git::https://github.com/marcinkubica/bosh-community-stemcell-ci-infra.git//terraform/concourse-infra?ref=terraform_v2"
  source = "../../../terraform-modules/concourse-infra"

  project = var.project
  region  = var.region
  zone    = var.zone


  #concourse_url = var.concourse_url

#   dns_address = {
#     name = "concourse-app-runtime-interfaces-ci-cloudfoundry-org"
#     url =  "https://concourse.app-runtime-interfaces.ci.cloudfoundry.org"
#   }

# SQL
  sql_instance_name = var.sql_instance_name
  sql_instance_tier = var.sql_instance_tier
  sql_instance_disk_size = var.sql_instance_disk_size
  sql_instance_secondary_zone = var.sql_instance_secondary_zone
  sql_instance_backup_location = var.sql_instance_backup_location

# DNS RECORD
  dns_zone = var.dns_zone
  dns_domain = var.dns_domain
  dns_record = var.dns_record
  dns_name = var.dns_name

# GKE
  gke_name = var.gke_name
  gke_controlplane_version = var.gke_controlplane_version
  gke_services_ipv4_cidr_block = var.gke_services_ipv4_cidr_block
  gke_cluster_ipv4_cidr = var.gke_cluster_ipv4_cidr
  gke_master_ipv4_cidr_block = var.gke_master_ipv4_cidr_block
  gke_default_pool_machine_type = var.gke_default_pool_machine_type
  gke_workers_pool_machine_type = var.gke_workers_pool_machine_type

# NETWORKING
  subnet_cidr = var.subnet_cidr

# IAM
  wg_ci_human_account_permissions = var.wg_ci_human_account_permissions
  wg_ci_cnrm_service_account_permissions = var.wg_ci_cnrm_service_account_permissions

}


