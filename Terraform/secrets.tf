resource "google_secret_manager_secret" "secret" {
  project   = var.gcp_project_id
  secret_id = "memorystore-${random_id.suffix.hex}"

  labels = {
    label = "memorystore-secret"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic" {
  secret = google_secret_manager_secret.secret.id

  secret_data = google_redis_instance.cache.auth_string
}
