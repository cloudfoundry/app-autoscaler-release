data "google_secret_manager_secret_version" "credhub_encryption_key" {
  project = var.project
  secret = "${var.gke_name}-credhub-encryption-key"
}

resource "kubernetes_secret_v1" "credhub_encryption_key" {
  metadata {
    name      = "credhub-encryption-key"
    namespace = "concourse"
  }

  data = {
    password = data.google_secret_manager_secret_version.credhub_encryption_key.secret_data
  }
}



