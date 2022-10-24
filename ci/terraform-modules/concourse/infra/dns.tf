data "google_dns_managed_zone" "dns" {
  name    = var.dns_zone
  project = var.project
}


resource "google_compute_address" "concourse_app" {
  project      = var.project
  region       = var.region
  address_type = "EXTERNAL"
  # gcp ip addres name can't contain dots
  name = replace("${var.dns_record}-${var.dns_zone}-${var.dns_domain}", ".", "-")
}


resource "google_dns_record_set" "concourse" {
  managed_zone = data.google_dns_managed_zone.dns.name
  name         = "${var.dns_record}.${data.google_dns_managed_zone.dns.dns_name}"
  type         = "A"
  rrdatas      = [google_compute_address.concourse_app.address]
  ttl          = 300
  project      = var.project
}

