variable "project" { default = null }
variable "region" { default = null }
variable "zone" { default = null }
variable "secondary_zone" { default = null }

variable "gke_name" { default = null }

variable "sql_instance_name" { default = null }
variable "sql_instance_tier" { default = null }
variable "sql_instance_disk_size" { default = null }
variable "sql_instance_backup_location" { default = null }
variable "sql_instance_secondary_zone" { default = null }

variable "dns_record" { nullable = false }
variable "dns_zone" { nullable = false }
variable "dns_domain" { nullable = false }
variable "dns_name" { nullable = false }