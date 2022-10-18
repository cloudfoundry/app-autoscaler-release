#
# firewall rule not needed as there's open communication


# resource "google_compute_firewall" "gke_to_all_vms" {
#   allow {
#     protocol = "ah"
#   }

#   allow {
#     protocol = "esp"
#   }

#   allow {
#     protocol = "icmp"
#   }

#   allow {
#     protocol = "sctp"
#   }

#   allow {
#     protocol = "tcp"
#   }

#   allow {
#     protocol = "udp"
#   }

#   # TODO: confirm this is intended direction as description is suggesting egress
#   direction     = "INGRESS"
#   source_ranges = [var.gke.cluster_ipv4_cidr]
#   disabled      = "false"
#   name          = "gke-${google_container_cluster.wg_ci.name}-to-all-vms-on-network"
#   network       = google_compute_network.vpc.name
#   priority      = "1000"
#   project       = var.project

# }