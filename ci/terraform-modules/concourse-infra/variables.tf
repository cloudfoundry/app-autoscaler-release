variable "project" {
  type        = string
  description = "Your GCP project name."
  nullable = false
}

variable "region" {
  type        = string
  description = "Project region"
  nullable = false
}

variable "zone" {
    type = string
    description = "Project primary zone"
    nullable = false
}


variable "sql_instance_name" { nullable = false }
variable "sql_instance_secondary_zone" { nullable = false }
variable "sql_instance_backup_location" { nullable = false }
variable "sql_instance_tier" { nullable = false }
variable "sql_instance_disk_size" { nullable = false }

variable "dns_record" { nullable = false }
variable "dns_zone" { nullable = false }
variable "dns_domain" { nullable = false }
variable "dns_name" { nullable = false }

variable "subnet_cidr" { nullable = false }


variable "gke_name" { nullable = false }

variable "wg_ci_human_account_permissions" {
  nullable = false
}

variable "wg_ci_cnrm_service_account_permissions" {
  nullable = false
}


# variable "dns_address" {
#  type = map(string)
#  description = "Concourse instance dns record name (on gcp) and public url"
#  default = {
#     name = null
#     url  =  null
#  }
# }





# variable "gke_name" {
#   default = "wg-ci"
#     controlplane_version      = "1.23.8-gke.1900"
#     cluster_ipv4_cidr         = "10.104.0.0/14"
#     services_ipv4_cidr_block  = "10.108.0.0/20"
#     master_ipv4_cidr_block    = "172.16.0.32/28"
#     machine_type_default_pool = "e2-standard-4"
#     machine_type_workers_pool = "n2-standard-4"
#   }
# }

# variable "kube" {
#   type = map(any)
#   default = {
#     config  = "~/.kube/config"
#     # TODO: try to provide context dynamically by reading GKE
#     context = "gke_app-runtime-interfaces-wg_europe-west3-a_wg-ci"
#   }
# }