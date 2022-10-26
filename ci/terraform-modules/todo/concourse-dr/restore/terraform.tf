terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
  }


  backend "gcs" {
    bucket = "terraform-state-wg-ci"
    prefix = "terraform/state/concourse-dr-restore"
  }
}


provider "kubernetes" {
  config_path    = var.kube.config
  config_context = var.kube.context
}

