terraform {
  required_providers {
    concourse = {
      source = "terraform-provider-concourse/concourse"
    }
  }


  backend "gcs" {
    bucket = "terraform-state-wg-ci"
    prefix = "terraform/state/concourse-manage"
  }
}


provider "concourse" {
  target = var.concourse_app.fly_target
}