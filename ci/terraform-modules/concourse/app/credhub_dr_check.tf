# --- Credhub encryption key
data "google_secret_manager_secret_version" "credhub_encryption_key" {
  project = var.project
  secret = "${var.gke_name}-credhub-encryption-key"

  depends_on = [carvel_kapp.concourse_app]
}

# Save encryption key once app is deployed
data "kubernetes_secret_v1" "credhub_encryption_key" {
  metadata {
    name      = "credhub-encryption-key"
    namespace = "concourse"
  }
  binary_data = {
    password = ""
  }

  depends_on = [carvel_kapp.concourse_app]
}

module "assertion_encryption_key_identical" {
  source  = "Invicton-Labs/assertion/null"

  condition = data.google_secret_manager_secret_version.credhub_encryption_key.secret_data == base64decode(data.kubernetes_secret_v1.credhub_encryption_key.binary_data.password)

  // The error message to print out if the condition evaluates to FALSE
  error_message = "*** Encryption keys in terraform and kubernetes do not match ***"

  depends_on = [carvel_kapp.concourse_app]
}
