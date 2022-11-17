resource "concourse_team" "app_autoscaler" {
  team_name = "app-autoscaler"

  owners = [
    "group:github:sap-cloudfoundry:app-autoscaler",
    "group:github:cloudfoundry:wg-app-runtime-interfaces-autoscaler-approvers"
  ]
}


data "concourse_teams" "teams" {
}

output "team_names" {
  value = data.concourse_teams.teams.names
}