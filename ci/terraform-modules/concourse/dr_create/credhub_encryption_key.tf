data "kubernetes_secret_v1" "credhub_encryption_key" {
  metadata {
    name      = "credhub-encryption-key"
    namespace = "concourse"
  }
  binary_data = {
    password = ""
  }
}

resource "google_secret_manager_secret" "credhub_encryption_key" {
  secret_id = "${var.gke_name}-credhub-encryption-key"
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

  depends_on = [data.kubernetes_secret_v1.credhub_encryption_key]
}


resource "google_secret_manager_secret_version" "credhub_encryption_key" {
  secret      = google_secret_manager_secret.credhub_encryption_key.id
  secret_data = base64decode(data.kubernetes_secret_v1.credhub_encryption_key.binary_data.password)
  lifecycle {
    prevent_destroy = true

    # If omitted or unset terraform destroys previous versions which will make it impossible to
    # restore them. This is relevant in case of a desaster recovery where the
    # history of secret might be needed to restore all credhub secrets.
    #
    # See: https://github.com/hashicorp/terraform-provider-google/issues/8653
    # Terraform will retrun error if user will attempt to create new version of credhub encryption key.
    # Such scenario should be a red flag and performed only when fully accepting the consequences of creating new
    #   secret version with terraform - and that is a permanent destruction of the previously saved secret.
    # In case new encryption key is to be saved please use 'terraform/terragrung state rm' to remove the version from the state.

  }

  depends_on = [data.kubernetes_secret_v1.credhub_encryption_key]
}

module "assertion_encryption_key_identical" {
  source  = "Invicton-Labs/assertion/null"

  condition = google_secret_manager_secret_version.credhub_encryption_key.secret_data == base64decode(data.kubernetes_secret_v1.credhub_encryption_key.binary_data.password)

  // The error message to print out if the condition evaluates to FALSE
  error_message = "*** Encryption keys in GCP Secret Manager and kubernetes secrets do not match ***"
}
