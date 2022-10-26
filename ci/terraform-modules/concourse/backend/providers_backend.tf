terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
    carvel = {
      source = "vmware-tanzu/carvel"
    }

  }
}

data "google_client_config" "provider" {}

data "google_container_cluster" "wg_ci" {
  project  = var.project
  name     = var.gke_name
  location = var.zone
}

provider "carvel" {
  kapp {
    kubeconfig {
      server = "https://${data.google_container_cluster.wg_ci.endpoint}"
      token = data.google_client_config.provider.access_token
      ca_cert = base64decode(
    data.google_container_cluster.wg_ci.master_auth[0].cluster_ca_certificate,
  )

    }
  }
}