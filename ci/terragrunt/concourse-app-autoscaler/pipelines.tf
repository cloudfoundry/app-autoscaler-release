resource "concourse_pipeline" "api_tester" {
  team_name     = "app-autoscaler"
  pipeline_name = "api-tester"

  is_exposed = false
  is_paused  = false

  pipeline_config        = file("pipelines/api-tester/pipeline.yml")
  pipeline_config_format = "yaml"
}

resource "concourse_pipeline" "cf_infrastructure" {
  team_name     = "app-autoscaler"
  pipeline_name = "cf-infrastructure"

  is_exposed = false
  is_paused  = false

  pipeline_config        = file("../../cf-infrastructure/pipeline.yml")
  pipeline_config_format = "yaml"
}

#waiting for vars feature to be available to public
#https://github.com/alphagov/terraform-provider-concourse/pull/50#issuecomment-1305426732
# resource "concourse_pipeline" "app_autoscaler_release" {
#   team_name     = "app-autoscaler"
#   pipeline_name = "app-autoscaler-release"

#   is_exposed = false
#   is_paused  = false

#   pipeline_config        = file("../../autoscaler/pipeline.yml")
#   pipeline_config_format = "yaml"
# #   vars = {
# #     branch = "main"
# #   }
# }
