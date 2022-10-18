resource "google_service_account" "gke_node_pools" {
  account_id   = "${var.gke.name}-node-pool"
  display_name = "Service account for ${var.gke.name} GKE node pools"
  project      = var.project
}


resource "google_container_node_pool" "default_pool" {
  cluster    = google_container_cluster.wg_ci.name
  node_count = "1"

  node_locations = [var.zone]
  project        = var.project
  location       = var.zone

  autoscaling {
    max_node_count       = "3"
    min_node_count       = "1"
    total_max_node_count = "0"
    total_min_node_count = "0"
  }


  management {
    auto_repair  = "true"
    auto_upgrade = "true"
  }

  max_pods_per_node = "110"
  name              = "default-pool"

  node_config {
    disk_size_gb    = "100"
    disk_type       = "pd-standard"
    image_type      = "COS_CONTAINERD"
    local_ssd_count = "0"
    machine_type    = var.gke.machine_type_default_pool

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/userinfo.email"
      ]
    preemptible     = "false"
    service_account = google_service_account.gke_node_pools.email

    shielded_instance_config {
      enable_integrity_monitoring = "true"
      enable_secure_boot          = "false"
    }

    spot = "false"

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  upgrade_settings {
    max_surge       = "1"
    max_unavailable = "0"
  }
}


resource "google_container_node_pool" "concourse_workers" {
  cluster    = google_container_cluster.wg_ci.name
  node_count = "2"

  node_locations = [var.zone]
  project        = var.project
  location       = var.zone

  autoscaling {
    max_node_count       = "4"
    min_node_count       = "2"
    total_max_node_count = "0"
    total_min_node_count = "0"
  }

  management {
    auto_repair  = "true"
    auto_upgrade = "true"
  }

  max_pods_per_node = "110"
  name              = "concourse-workers"

  node_config {
    disk_size_gb    = "100"
    disk_type       = "pd-standard"
    image_type      = "COS_CONTAINERD"
    local_ssd_count = "1"
    machine_type    = var.gke.machine_type_workers_pool

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/userinfo.email"]
    preemptible     = "false"
    service_account = google_service_account.gke_node_pools.email

    shielded_instance_config {
      enable_integrity_monitoring = "true"
      enable_secure_boot          = "false"
    }

    spot = "false"
    tags = ["workers"]

    taint {
      effect = "NO_SCHEDULE"
      key    = "workers"
      value  = "true"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }


  upgrade_settings {
    max_surge       = "1"
    max_unavailable = "0"
  }
}

