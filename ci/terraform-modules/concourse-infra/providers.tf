data "google_client_config" "provider" {}

provider kubernetes {
  host  = "https://${google_container_cluster.wg_ci.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    google_container_cluster.wg_ci.master_auth[0].cluster_ca_certificate,
    )
  }
