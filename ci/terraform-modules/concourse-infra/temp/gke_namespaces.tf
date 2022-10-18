resource "kubernetes_namespace" "concourse" {
  metadata {
    name = "concourse"
  }

  lifecycle {
    ignore_changes = all
  }

}