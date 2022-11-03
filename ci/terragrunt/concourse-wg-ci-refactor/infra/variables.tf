variable "project" { default = null }
variable "region" { default = null }
variable "zone" { default = null }
variable "secondary_zone" { default = null }

variable "sql_instance_name" { default = null }
variable "sql_instance_tier" { default = null }
variable "sql_instance_disk_size" { default = null }
variable "sql_instance_backup_location" { default = null }
variable "sql_instance_secondary_zone" { default = null }

variable "subnet_cidr" { nullable = false }

variable "dns_record" { nullable = false }
variable "dns_zone" { nullable = false }
variable "dns_domain" { nullable = false }
variable "dns_name" { nullable = false }

variable "gke_name" { default = null }
variable "gke_controlplane_version" { nullable = false }
variable "gke_cluster_ipv4_cidr" { nullable = false }
variable "gke_services_ipv4_cidr_block" { nullable = false }
variable "gke_master_ipv4_cidr_block" { nullable = false }
variable "gke_default_pool_machine_type" { nullable = false }
variable "gke_workers_pool_machine_type" { nullable = false }

variable "wg_ci_human_account_permissions" { nullable = false }
variable "wg_ci_cnrm_service_account_permissions" { nullable = false }

variable "github_secret_name" { nullable = false }
