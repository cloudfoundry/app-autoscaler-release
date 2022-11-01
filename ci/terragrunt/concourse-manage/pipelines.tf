resource "concourse_pipeline" "hello_world" {
  team_name     = "app-autoscaler"
  pipeline_name = "hello-world"

  is_exposed = true
  is_paused  = false

  pipeline_config        = file("pipelines/hello-world/pipeline.yml")
  pipeline_config_format = "yaml"
}

resource "concourse_pipeline" "api-tester" {
  team_name     = "app-autoscaler"
  pipeline_name = "api-tester"

  is_exposed = true
  is_paused  = false

  pipeline_config        = file("pipelines/api-tester/pipeline.yml")
  pipeline_config_format = "yaml"
}
