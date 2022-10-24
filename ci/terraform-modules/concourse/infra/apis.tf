# Set apis to not disable in case we issue `terraform destroy`
resource "google_project_service" "apis" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "secretmanager.googleapis.com",
    "sqladmin.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com"
  ])
  service            = each.key
  project            = var.project
  disable_on_destroy = false
}
