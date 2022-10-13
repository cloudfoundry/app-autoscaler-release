locals {
    config = yamldecode(file("../config.yaml"))
}


module "infra" {
#variables can't be used here
#source = "git::https://github.com/marcinkubica/bosh-community-stemcell-ci-infra.git//terraform/concourse-infra?ref=terraform"


source = "../../../../bosh-community-stemcell-ci-infra//terraform/concourse-infra/"

  project = local.config.project

 }

# output "test" {
#   #value = "${tomap(var.module_sources.concourse_infra.ref)}"
#   value = local.config.gke.name
#  }

 

