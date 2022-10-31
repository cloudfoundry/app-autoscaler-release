terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
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
  cluster_ca_certificate = base64decode(data.google_container_cluster.wg_ci.master_auth[0].cluster_ca_certificate)
}

