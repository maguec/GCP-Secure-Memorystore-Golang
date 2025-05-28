resource "google_secret_manager_secret" "secret-ip" {
  project   = var.gcp_project_id
  secret_id = "valkey-${random_id.suffix.hex}-ip"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-ip" {
  secret      = google_secret_manager_secret.secret-ip.id
  secret_data = google_memorystore_instance.cache.discovery_endpoints[0].address
}

resource "google_secret_manager_secret" "secret-cert" {
  project   = var.gcp_project_id
  secret_id = "valkey-${random_id.suffix.hex}-cert"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

# This is manual no #TODO: fix when the terraform provider gets fixed
#resource "google_secret_manager_secret_version" "secret-version-basic-cert" {
#  secret      = google_secret_manager_secret.secret-cert.id
#  secret_data = google_memorystore_instance.cache.server_ca_certs[0].cert
#}
#
resource "google_secret_manager_secret" "secret-port" {
  project   = var.gcp_project_id
  secret_id = "valkey-${random_id.suffix.hex}-port"
  labels    = { label = "memorystore-secret" }
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic-port" {
  secret      = google_secret_manager_secret.secret-port.id
  secret_data = google_memorystore_instance.cache.discovery_endpoints[0].port
}
