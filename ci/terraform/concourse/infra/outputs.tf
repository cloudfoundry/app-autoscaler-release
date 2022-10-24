# output "load_balancer_ip" {
#   # Consumed by concourse-app helm chart for Concourse
#   value = module.concourse-infra.google_compute_address.concourse_app.address
# }

# output "github_oauth" {
#   # Consumed by Concourse github auth
#   value     = module.concourse-infra.google_secret_manager_secret.github_oauth
#   sensitive = true
# }

