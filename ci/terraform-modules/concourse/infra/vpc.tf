resource "google_compute_network" "vpc" {
  name                    = "${var.gke_name}-vpc"
  project                 = var.project
  auto_create_subnetworks = "false"
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name                     = "${var.gke_name}-subnet"
  region                   = var.region
  project                  = var.project
  network                  = google_compute_network.vpc.name
  private_ip_google_access = true
  ip_cidr_range            = var.gke_subnet_cidr
}

# data "google_compute_network" "vpc" {
#   name                            = var.vpc_name
#   project                         = var.project
# }

# data "google_compute_subnetwork" "subnet" {
#   name                     = var.subnet_name
#   project                  = var.project
#   region                   = var.region
# }


# TODO: determine if this is only bosh stemcell infra requirement
# # create subnets required for deploying bosh workloads and running  tests
# resource "google_compute_subnetwork" "bosh_integration_0" {
#   ip_cidr_range              = "10.100.0.0/24"
#   name                       = "bosh-integration-0"
#   network                    = google_compute_network.vpc.name
#   private_ip_google_access   = "false"
#   private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
#   project                    = var.project
#   purpose                    = "PRIVATE"
#   region                     = var.region
#   stack_type                 = "IPV4_ONLY"
# }

# resource "google_compute_subnetwork" "bosh_integration_1" {
#   ip_cidr_range              = "10.100.1.0/24"
#   name                       = "bosh-integration-1"
#   network                    = google_compute_network.vpc.name
#   private_ip_google_access   = "false"
#   private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
#   project                    = var.project
#   purpose                    = "PRIVATE"
#   region                     = var.region
#   stack_type                 = "IPV4_ONLY"
# }

# resource "google_compute_subnetwork" "bosh_integration_2" {
#   ip_cidr_range              = "10.100.2.0/24"
#   name                       = "bosh-integration-2"
#   network                    = google_compute_network.vpc.name
#   private_ip_google_access   = "false"
#   private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
#   project                    = var.project
#   purpose                    = "PRIVATE"
#   region                     = var.region
#   stack_type                 = "IPV4_ONLY"
# }

# resource "google_compute_subnetwork" "bosh_integration_3" {
#   ip_cidr_range              = "10.100.3.0/24"
#   name                       = "bosh-integration-3"
#   network                    = google_compute_network.vpc.name
#   private_ip_google_access   = "false"
#   private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
#   project                    = var.project
#   purpose                    = "PRIVATE"
#   region                     = var.region
#   stack_type                 = "IPV4_ONLY"
# }
