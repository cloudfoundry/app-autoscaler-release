resource "google_project_iam_custom_role" "wg_ci_role" {
  permissions = toset(yamldecode(var.wg_ci_human_account_permissions))

  project     = var.project
  role_id     = "${replace(var.gke_name, "-", "_")}WgCiCustomRole"
  stage       = "GA"
  title       = "WG CI Manage [${var.gke_name}]"
  description = "Permissions for humans to manage ${var.gke_name} gke-cluster"
}

