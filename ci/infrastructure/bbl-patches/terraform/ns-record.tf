variable "parent_dns_zone" {
    type = string
    description = "Parent autoscaler zone name to use for the recordset."
}

data "google_dns_managed_zone" "parent_dns_zone" {
  name    = var.parent_dns_zone
  project = var.project
}

resource "google_dns_record_set" "autoscaler_app_runtime_interfaces" {
  name       = ".${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = [google_dns_managed_zone.env_dns_zone]
  type       = "NS"
  ttl        = 300

  rrdatas     = tolist(data.google_dns_managed_zone.env_dns_zone.name_servers)

  managed_zone = "${data.google_dns_managed_zone.parent_dns_zone.name}"
  project = var.project
}
