resource "google_sql_database_instance" "concourse" {
  database_version = "POSTGRES_13"
  name             = var.sql_instance_name
  project          = var.project
  region           = var.region

  # recommended protection via GCP SQL Instance settings
  # https://console.cloud.google.com/sql/instances/ -> select instance name -> edit
  # ->  Data Protection -> tick: Enable delete protection
  deletion_protection = false

  settings {
    activation_policy = "ALWAYS"
    availability_type = "REGIONAL"

    backup_configuration {

      location = var.sql_instance_backup_location
      backup_retention_settings {
        retained_backups = "7"
        retention_unit   = "COUNT"
      }

      binary_log_enabled             = "false"
      enabled                        = "true"
      point_in_time_recovery_enabled = "true"
      start_time                     = "00:00"
      transaction_log_retention_days = "7"
    }

    disk_autoresize       = "true"
    disk_autoresize_limit = "0"
    disk_size             = var.sql_instance_disk_size
    disk_type             = "PD_SSD"

    ip_configuration {
      ipv4_enabled = "true"
      require_ssl  = "false"
    }

    location_preference {
      zone           = var.zone
      secondary_zone = var.sql_instance_secondary_zone
    }

    maintenance_window {
      day          = 7 #Sunday
      hour         = 0 #0:00 - 1:00 hours
      update_track = "stable"
    }

    pricing_plan = "PER_USE"
    tier         = var.sql_instance_tier

  }
}

