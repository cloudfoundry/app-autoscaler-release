resource "concourse_pipeline" "api-tester" {
  team_name     = "app-autoscaler"
  pipeline_name = "api-tester"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("pipelines/api-tester/pipeline.yml")
  pipeline_config_format = "yaml"
}

resource "concourse_pipeline" "infrastructure" {
  team_name     = "app-autoscaler"
  pipeline_name = "infrastructure"

  is_exposed = true
  is_paused  = true

  pipeline_config        = file("../../infrastructure/pipeline.yml")
  pipeline_config_format = "yaml"
}
