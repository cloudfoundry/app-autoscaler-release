package main

deny[msg] {
  ig := input.instance_groups[_]
  ig.name == "postgres_autoscaler"
  not ig.persistent_disk_type

  msg := "Persistent Disk Type is not configured by default"
}
