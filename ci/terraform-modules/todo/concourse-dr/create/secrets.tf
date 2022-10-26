# Save encryption key once app is deployed
data "kubernetes_secret_v1" "credhub_encryption_key" {
  metadata {
    name      = "credhub-encryption-key"
    namespace = var.concourse_app.namespace
  }
  binary_data = {
    password = ""
  }
}

resource "google_secret_manager_secret" "credhub_encryption_key" {
  secret_id = "${var.gke.name}-credhub-encryption-key"
  project   = var.project
  # when creating versions with gcloud it creates empty labels
  labels = {
  }
  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}


resource "google_secret_manager_secret_version" "credhub_encryption_key" {
  secret      = google_secret_manager_secret.credhub_encryption_key.id
  secret_data = base64decode(data.kubernetes_secret_v1.credhub_encryption_key.binary_data.password)
  lifecycle {
    # If omitted or unset terraform destroys previous versions which will make it impossible to
    # restore them. This is relevant in case of a desaster recovery where the
    # history of secret might be needed to restore all credhub secrets.
    #
    # See: https://github.com/hashicorp/terraform-provider-google/issues/8653
    prevent_destroy = true
    # no further changes will be applied 
    # if the encryption key will change it will not be updated on secret manager
    ignore_changes = all

  }
}

