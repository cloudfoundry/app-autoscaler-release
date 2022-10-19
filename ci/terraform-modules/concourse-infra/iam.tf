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
    "projects/${var.project}/roles/${google_project_iam_custom_role.wg_ci_cnrm.role_id}"
  ])
  role = each.key

  depends_on = [
    google_project_iam_custom_role.wg_ci_cnrm
  ]
}

resource "google_service_account_iam_member" "cnrm_system" {
  service_account_id = google_service_account.cnrm_system.id
  member             = "serviceAccount:${var.project}.svc.id.goog[cnrm-system/cnrm-controller-manager]"
  role               = "roles/iam.workloadIdentityUser"
}

resource "google_project_iam_custom_role" "wg_ci_role" {
  permissions = [
    "container.clusterRoles.bind",
    "container.clusterRoles.create",
    "container.clusterRoles.delete",
    "container.clusterRoles.escalate",
    "container.clusterRoles.get",
    "container.clusterRoles.list",
    "container.clusterRoles.update",
    "container.clusterRoleBindings.create",
    "container.clusterRoleBindings.delete",
    "container.clusterRoleBindings.get",
    "container.clusterRoleBindings.list",
    "container.clusterRoleBindings.update",
    "container.configMaps.get",

    "iam.roles.create",
    "iam.roles.update",

    "iam.serviceAccounts.setIamPolicy",

    "resourcemanager.projects.get",
    "resourcemanager.projects.getIamPolicy",
    "resourcemanager.projects.setIamPolicy",
    # TODO: rather give access to particular secret
    "secretmanager.versions.access"
    ]

  project = var.project
  role_id = "${replace(var.gke_name, "-", "_")}WgCiCustomRole"
  stage   = "GA"
  title   = "WG CI Manage [${var.gke_name}]"
  description = "Permissions for humans to manage ${var.gke_name} gke-cluster"
}

resource "google_project_iam_custom_role" "wg_ci_cnrm" {
  permissions = [
    "cloudsql.users.create",
    "cloudsql.users.delete",
    "cloudsql.users.get",
    "cloudsql.users.list",
    "cloudsql.users.update",
    "cloudsql.databases.get",
    "cloudsql.databases.list",
    "cloudsql.databases.update"
  ]

  project = var.project
  role_id = "${replace(var.gke_name, "-", "_")}WgCiCNRMcustomRole"
  stage   = "GA"
  title   = "WG CI CNRM-SYSTEM [${var.gke_name}]"
  description = "Additional permissions for cnrm-system on ${var.gke_name} Concourse deployment"
}
