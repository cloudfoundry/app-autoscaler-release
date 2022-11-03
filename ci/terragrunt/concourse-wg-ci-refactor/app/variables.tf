variable "project" { default = null }
variable "region" { default = null }
variable "zone" { default = null }

variable "gke_name" { default = null }

variable "concourse_helm_version" { nullable = false }

variable "concourse_github_mainTeam" { nullable = false }
variable "concourse_github_mainTeamUser" { nullable = false }


variable "load_balancer_ip" { nullable = false }
variable "load_balancer_dns" { nullable = false }
variable "github_oauth_secret_name" { nullable = false }
