resource "google_secret_manager_secret" "github_oauth" {
  secret_id = var.github_secret_name
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
