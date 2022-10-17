
module "concourse-infra" {
  #variables can't be used here
  #source = "git::https://github.com/marcinkubica/bosh-community-stemcell-ci-infra.git//terraform/concourse-infra?ref=terraform_v2"
  source = "../../../../bosh-community-stemcell-ci-infra//terraform/concourse-infra/"

  project = var.project
  region  = var.region
  zone    = var.zone


  #concourse_url = var.concourse_url

#   dns_address = {
#     name = "concourse-app-runtime-interfaces-ci-cloudfoundry-org"
#     url =  "https://concourse.app-runtime-interfaces.ci.cloudfoundry.org"
#   }

# SQL
  sql_instance_name = "${var.gke_name}-concourse"
  sql_instance_tier = var.sql_instance_tier
  sql_instance_disk_size = var.sql_instance_disk_size
  sql_instance_secondary_zone = var.sql_instance_secondary_zone
  sql_instance_backup_location = var.sql_instance_backup_location


# GKE
# gke_name = var.base_name
#    gke = {
#      name = "wg-ci"
#      controlplane_version = "1.23.8-gke.1900"
    #  cluster_ipv4_cidr = ""
    #  services_ipv4_cidr_block  = ""
    #  #master_ipv4_cidr_block    = ""
    #  machine_type_default_pool = ""
    #  machine_type_workers_pool = ""
 }
