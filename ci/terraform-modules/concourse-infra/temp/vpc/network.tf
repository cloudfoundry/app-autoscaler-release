resource "google_compute_network" "vpc" {
  auto_create_subnetworks         = "true"
  delete_default_routes_on_create = "false"
  description                     = "Default network for the project"
  enable_ula_internal_ipv6        = "false"
  mtu                             = "0"
  name                            = "default"
  project                         = var.project
  routing_mode                    = "REGIONAL"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_compute_subnetwork" "default" {
  ip_cidr_range            = "10.156.0.0/20"
  name                     = "default"
  network                  = google_compute_network.vpc.name
  private_ip_google_access = "true"
  project                  = var.project
  purpose                  = "PRIVATE"
  region                   = var.region
  stack_type               = "IPV4_ONLY"
  lifecycle {
    prevent_destroy = true
  }
}

# create subnets required for deploying bosh workloads and running  tests
resource "google_compute_subnetwork" "bosh_integration_0" {
  ip_cidr_range              = "10.100.0.0/24"
  name                       = "bosh-integration-0"
  network                    = google_compute_network.vpc.name
  private_ip_google_access   = "false"
  private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
  project                    = var.project
  purpose                    = "PRIVATE"
  region                     = var.region
  stack_type                 = "IPV4_ONLY"
}

resource "google_compute_subnetwork" "bosh_integration_1" {
  ip_cidr_range              = "10.100.1.0/24"
  name                       = "bosh-integration-1"
  network                    = google_compute_network.vpc.name
  private_ip_google_access   = "false"
  private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
  project                    = var.project
  purpose                    = "PRIVATE"
  region                     = var.region
  stack_type                 = "IPV4_ONLY"
}

resource "google_compute_subnetwork" "bosh_integration_2" {
  ip_cidr_range              = "10.100.2.0/24"
  name                       = "bosh-integration-2"
  network                    = google_compute_network.vpc.name
  private_ip_google_access   = "false"
  private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
  project                    = var.project
  purpose                    = "PRIVATE"
  region                     = var.region
  stack_type                 = "IPV4_ONLY"
}

resource "google_compute_subnetwork" "bosh_integration_3" {
  ip_cidr_range              = "10.100.3.0/24"
  name                       = "bosh-integration-3"
  network                    = google_compute_network.vpc.name
  private_ip_google_access   = "false"
  private_ipv6_google_access = "DISABLE_GOOGLE_ACCESS"
  project                    = var.project
  purpose                    = "PRIVATE"
  region                     = var.region
  stack_type                 = "IPV4_ONLY"
}
