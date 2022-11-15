variable "project" { default = null }
variable "zone" { default = null }
variable "region" { default = null }
variable "sql_instance_name" { default = null }

variable "gke_name" { nullable = false }
variable "wg_ci_cnrm_service_account_permissions" { nullable = false }