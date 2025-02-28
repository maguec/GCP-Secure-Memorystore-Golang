resource "google_secret_manager_secret" "secret-auth" {
  project   = var.gcp_project_id
  secret_id = "memorystore-${random_id.suffix.hex}-auth"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-auth" {
  secret      = google_secret_manager_secret.secret-auth.id
  secret_data = google_redis_instance.cache.auth_string
}

resource "google_secret_manager_secret" "secret-ip" {
  project   = var.gcp_project_id
  secret_id = "memorystore-${random_id.suffix.hex}-ip"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-ip" {
  secret      = google_secret_manager_secret.secret-ip.id
  secret_data = google_redis_instance.cache.host
}

resource "google_secret_manager_secret" "secret-cert" {
  project   = var.gcp_project_id
  secret_id = "memorystore-${random_id.suffix.hex}-cert"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-cert" {
  secret      = google_secret_manager_secret.secret-cert.id
  secret_data = google_redis_instance.cache.server_ca_certs[0].cert
}

resource "google_secret_manager_secret" "secret-port" {
  project   = var.gcp_project_id
  secret_id = "memorystore-${random_id.suffix.hex}-port"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-port" {
  secret      = google_secret_manager_secret.secret-port.id
  secret_data = google_redis_instance.cache.port
}
