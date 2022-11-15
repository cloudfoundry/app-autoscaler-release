output "concourse_url" {
    value = "Your concourse instance is available at https://${var.load_balancer_dns}"
}