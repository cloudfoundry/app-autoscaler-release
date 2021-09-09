package main

postgres_has_persistent_disk {
  ig := input.instance_groups[_]

  ig.name == "postgres_autoscaler"
  ig.persistent_disk_type == "10GB"
}

deny[msg] {
  not postgres_has_persistent_disk
  msg := "persistent disk for postgres should be configured"
}
