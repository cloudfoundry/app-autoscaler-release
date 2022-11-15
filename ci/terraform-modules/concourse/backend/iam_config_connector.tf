resource "google_service_account" "cnrm_system" {
  account_id  = "${var.gke_name}-cnrm-system"
  description = "Config Connector account for ${var.gke_name} GKE"
  disabled    = "false"
  project     = var.project
}


resource "google_project_iam_member" "cnrm_system" {
  project = var.project
  member  = "serviceAccount:${google_service_account.cnrm_system.email}"
  for_each = toset([
    "roles/resourcemanager.projectIamAdmin",
    "roles/iam.serviceAccountAdmin",
    "roles/cloudsql.instanceUser",
    "projects/${var.project}/roles/${google_project_iam_custom_role.wg_ci_cnrm.role_id}"
  ])
  role = each.key

  depends_on = [ google_project_iam_custom_role.wg_ci_cnrm ]
}

resource "google_service_account_iam_member" "cnrm_system" {
  service_account_id = google_service_account.cnrm_system.id
  member             = "serviceAccount:${var.project}.svc.id.goog[cnrm-system/cnrm-controller-manager]"
  role               = "roles/iam.workloadIdentityUser"
}

resource "google_project_iam_custom_role" "wg_ci_cnrm" {
  permissions = toset(yamldecode(var.wg_ci_cnrm_service_account_permissions))

  project     = var.project
  role_id     = "${replace(var.gke_name, "-", "_")}WgCiCNRMcustomRole"
  stage       = "GA"
  title       = "WG CI CNRM-SYSTEM [${var.gke_name}]"
  description = "Additional permissions for cnrm-system on ${var.gke_name} Concourse deployment"
}

