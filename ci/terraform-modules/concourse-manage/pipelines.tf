resource "concourse_pipeline" "hello_world" {
  team_name     = "app-autoscaler"
  pipeline_name = "hello-world"

  is_exposed = true
  is_paused  = false

  pipeline_config        = file("hello-world-pipeline.yml")
  pipeline_config_format = "yaml"
}
