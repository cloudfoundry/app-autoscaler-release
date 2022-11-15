variable "project" {
  type        = string
  description = "Your GCP project name."
  nullable    = false
}

variable "region" {
  type        = string
  description = "Project region"
  nullable    = false
}

variable "zone" {
  type        = string
  description = "Project primary zone"
  nullable    = false
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

variable "wg_ci_human_account_permissions" { nullable = false }

variable "gke_name" { nullable = false }
variable "gke_controlplane_version" { nullable = false }
variable "gke_cluster_ipv4_cidr" { nullable = false }
variable "gke_services_ipv4_cidr_block" { nullable = false }
variable "gke_master_ipv4_cidr_block" { nullable = false }

variable "gke_default_pool_machine_type" { nullable = false }
variable "gke_default_pool_node_count" { nullable = false }
variable "gke_default_pool_autoscaling_max" { nullable = false }
variable "gke_default_pool_ssd_count" { nullable = false }

variable "gke_workers_pool_machine_type" { nullable = false }
variable "gke_workers_pool_node_count" { nullable = false }
variable "gke_workers_pool_autoscaling_max" { nullable = false }
variable "gke_workers_pool_ssd_count" { nullable = false }

variable "gke_subnet_cidr" { nullable = false }
variable "gke_cloud_nat_min_ports_per_vm" { nullable = false }

variable "github_secret_name" { nullable = false }

