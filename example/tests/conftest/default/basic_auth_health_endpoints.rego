package main

contains_variable[name] {
  v := input.variables[_]
  v.type == "password" 
  name := v.name
}

deny[msg] {
  not contains_variable["autoscaler_eventgenerator_health_password"]
  msg := "variable autoscaler_eventgenerator_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_metricsforwarder_health_password"]
  msg := "variable autoscaler_metricsforwarder_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_metricsgateway_health_password"]
  msg := "variable autoscaler_metricsgateway_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_metricsserver_health_password"]
  msg := "variable autoscaler_metricsserver_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_scalingengine_health_password"]
  msg := "variable autoscaler_scalingengine_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_operator_health_password"]
  msg := "variable autoscaler_operator_health_password is missing"
}

deny[msg] {
  not contains_variable["autoscaler_scheduler_health_password"]
  msg := "variable autoscaler_scheduler_health_password is missing"
}
