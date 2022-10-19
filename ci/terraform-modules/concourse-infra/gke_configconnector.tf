resource "kubernetes_manifest" "config_connector" {
  manifest = {
    apiVersion = "core.cnrm.cloud.google.com/v1beta1"
    kind       = "ConfigConnector"
    metadata = {
      "name" = "configconnector.core.cnrm.cloud.google.com"
    }
    spec = {
      mode                 = "cluster"
      googleServiceAccount = "${google_service_account.cnrm_system.email}"

    }
  }
  depends_on = [ google_container_cluster.wg_ci ]
}

