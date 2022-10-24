resource "google_secret_manager_secret" "github_oauth" {
  secret_id = "${var.gke_name}-concourse-github-oauth"
  project   = var.project

  # when creating versions with gcloud it creates empty labels
  labels = {

  }
  replication {
    user_managed {
      replicas {
        location = "europe-west3"
      }
    }
  }
}
