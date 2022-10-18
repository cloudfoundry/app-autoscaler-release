resource "google_compute_router" "nat_router" {
  encrypted_interconnect_router = "false"
  name                          = "nat-router"
  network                       = google_compute_network.vpc.name
  project                       = var.project
  region                        = var.region
}


resource "google_compute_router_nat" "nat_config" {
  name                                = "nat-config"
  router                              = google_compute_router.nat_router.name
  region                              = google_compute_router.nat_router.region
  nat_ip_allocate_option              = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat  = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  min_ports_per_vm                    = "4095"
  enable_endpoint_independent_mapping = false
  tcp_established_idle_timeout_sec    = 60

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }

  depends_on = [google_compute_router.nat_router]
}