output "load_balancer_ip" {
  # Consumed by Concourse app helm chart
  value = google_compute_address.concourse_app.address
}

output "load_balancer_dns" {
  # Consumed by concourse-app helm chart
  value = trimsuffix(google_dns_record_set.concourse.name, ".")
}
