resource "google_container_cluster" "wg_ci" {
  # beta provider uset to allow enable of Config Connector
  provider                 = google-beta
  name                     = var.gke_name
  location                 = var.zone
  project                  = var.project
  initial_node_count       = "1"
  remove_default_node_pool = true
  min_master_version       = var.gke_controlplane_version

  release_channel {
    channel = "STABLE"
  }

  workload_identity_config {
    workload_pool = "${var.project}.svc.id.goog"
  }

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
  }

  cluster_autoscaling {
    enabled = "false"
  }

  ip_allocation_policy {
    cluster_ipv4_cidr_block  = var.gke_cluster_ipv4_cidr
    services_ipv4_cidr_block = var.gke_services_ipv4_cidr_block
  }
  private_cluster_config {
    enable_private_endpoint = "false"
    enable_private_nodes    = "true"

    master_global_access_config {
      enabled = "false"
    }

    master_ipv4_cidr_block = var.gke_master_ipv4_cidr_block
  }

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
  network_policy {
    enabled  = "false"
    provider = "PROVIDER_UNSPECIFIED"
  }

  networking_mode = "VPC_NATIVE"

  # other config
  addons_config {
    config_connector_config {
      enabled = "true"
    }
    gce_persistent_disk_csi_driver_config {
      enabled = "true"
    }

    horizontal_pod_autoscaling {
      disabled = "true"
    }

    http_load_balancing {
      disabled = "true"
    }

    network_policy_config {
      disabled = "true"
    }
  }

  database_encryption {
    state = "DECRYPTED"
  }

  default_max_pods_per_node = "110"

  default_snat_status {
    disabled = "true"
  }

  binary_authorization {
    evaluation_mode = "DISABLED"
  }

  enable_intranode_visibility = "false"
  enable_kubernetes_alpha     = "false"
  enable_legacy_abac          = "false"
  enable_shielded_nodes       = "true"
  enable_tpu                  = "false"

  master_auth {
    client_certificate_config {
      issue_client_certificate = "false"
    }
  }


  service_external_ips_config {
    enabled = "true"
  }

}
