# using kubectl_manifest as we can't modify existing CRD created by Connfig Connector with kubernetes_manifest
resource "kubernetes_namespace" "concourse" {
  metadata {
    name = "concourse"
  }

  lifecycle {
    ignore_changes = [metadata]
  }

  depends_on = [data.google_container_cluster.wg_ci]
}


resource "kubectl_manifest" "config_connector" {
  yaml_body = <<YAML
    apiVersion: core.cnrm.cloud.google.com/v1beta1
    kind: ConfigConnector
    metadata:
     name: configconnector.core.cnrm.cloud.google.com
    spec:
      mode: cluster
      googleServiceAccount: ${google_service_account.cnrm_system.email}
  YAML

  depends_on = [data.google_container_cluster.wg_ci, kubernetes_namespace.concourse]
}


