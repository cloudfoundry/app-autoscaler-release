variable "project" { nullable = false }
variable "region" { nullable = false }
variable "zone" { nullable = false }

variable "gke_name" { nullable = false }

variable "concourse_helm_version" { nullable = false }

variable "concourse_github_mainTeam" { nullable = false }
variable "concourse_github_mainTeamUser" { nullable = false }

variable "load_balancer_ip" { nullable = false }
variable "load_balancer_dns" { nullable = false }
variable "github_oauth_secret_name" { nullable = false }
