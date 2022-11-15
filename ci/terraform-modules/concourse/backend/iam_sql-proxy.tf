resource "google_service_account" "sql_proxy" {
  account_id  = "${var.gke_name}-sql-proxy"
  display_name = "Used by Cloud SQL Auth proxy [${var.gke_name}]"
  disabled    = "false"
  project     = var.project
}

resource "google_service_account_iam_member" "sql_proxy" {
  service_account_id = google_service_account.sql_proxy.id
  member             = "serviceAccount:${var.project}.svc.id.goog[concourse/${var.gke_name}-sql-proxy]"
  role               = "roles/iam.workloadIdentityUser"
}

resource "google_project_iam_member" "sql_proxy" {
  project = var.project
  member  = "serviceAccount:${google_service_account.sql_proxy.email}"
  role = "roles/cloudsql.client"

}

resource "kubectl_manifest" "sql_proxy_service_account" {
  yaml_body = <<YAML
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: ${var.gke_name}-sql-proxy
     namespace: concourse
     annotations:
       iam.gke.io/gcp-service-account: ${google_service_account.sql_proxy.email}
  YAML

  depends_on = [data.google_container_cluster.wg_ci, google_service_account.sql_proxy, kubectl_manifest.config_connector, kubernetes_namespace.concourse ]
}