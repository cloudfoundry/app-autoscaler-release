terraform {
  required_providers {
    kubectl = {
      source = "gavinbunney/kubectl"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}

data "google_client_config" "provider" {}

data "google_container_cluster" "wg_ci" {
  project  = var.project
  name     = var.gke_name
  location = var.zone
}

provider "kubernetes" {
  host  = "https://${data.google_container_cluster.wg_ci.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    google_container_cluster.wg_ci.master_auth[0].cluster_ca_certificate,
  )
}

provider "kubectl" {
  host  = "https://${data.google_container_cluster.wg_ci.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    google_container_cluster.wg_ci.master_auth[0].cluster_ca_certificate,
  )
  load_config_file = false

}