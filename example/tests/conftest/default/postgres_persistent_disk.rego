package main

postgres_has_persistent_disk {
  ig := input.instance_groups[_]

  ig.name == "postgres_autoscaler"
  ig.persistent_disk_type
}

deny[msg] {
  postgres_has_persistent_disk
  msg := "Persistent Disk Type is not configured by default"
}
