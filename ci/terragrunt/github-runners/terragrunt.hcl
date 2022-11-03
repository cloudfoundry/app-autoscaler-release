remote_state {
  backend = "gcs"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite"
  }
  config = {
    bucket         = "terraform-app-autoscaler"
    prefix         = "github-runners"
    project        = "app-runtime-interfaces-wg"
    location       = "europe-west3"
    enable_bucket_policy_only = true
  }
}

