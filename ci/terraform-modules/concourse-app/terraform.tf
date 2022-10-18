terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }

    carvel = {
      source = "vmware-tanzu/carvel"
    }
  }


  backend "gcs" {
    bucket = "terraform-state-wg-ci"
    prefix = "terraform/state/concourse-app"
  }
}

data "terraform_remote_state" "infra" {
  backend = "gcs"
  config = {
    bucket = "terraform-state-wg-ci"
    prefix = "terraform/state/infra"

  }
}

provider "google-beta" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "gke_app-runtime-interfaces-wg_europe-west3-a_wg-ci"
}

provider "carvel" {
  kapp {
    kubeconfig {
      from_env = true
      #kubeconfig = var.kube.config
      context = var.kube.context
    }
  }
}

