resource "google_redis_instance" "cache" {
  project                 = var.gcp_project_id
  region                  = var.gcp_region
  name                    = "memorystore-${random_id.suffix.hex}"
  memory_size_gb          = 1
  auth_enabled            = true
  transit_encryption_mode = "SERVER_AUTHENTICATION"
  authorized_network      = google_compute_network.vpc.name
}
