package main

deny[msg] {
  ig := input.instance_groups[_]
  ig.name == "postgres_autoscaler"
  ig.persistent_disk_type == "10GB"

  msg := "Persistent Disk Type should be set to 10GB"
}
