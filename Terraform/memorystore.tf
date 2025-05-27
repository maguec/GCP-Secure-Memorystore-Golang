resource "google_memorystore_instance" "cache" {
  project        = var.gcp_project_id
  location       = var.gcp_region
  instance_id    = "valkey-${random_id.suffix.hex}"
  shard_count    = 3
  engine_version = "VALKEY_8_0"
  mode           = "CLUSTER"

  desired_psc_auto_connections {
    network    = "projects/${var.gcp_project_id}/global/networks/${google_compute_network.vpc.name}"
    project_id = var.gcp_project_id
  }

  replica_count           = 0
  node_type               = "STANDARD_SMALL"
  transit_encryption_mode = "SERVER_AUTHENTICATION"
  authorization_mode      = "IAM_AUTH"

  depends_on = [
    google_network_connectivity_service_connection_policy.service_connection_policies,
  ]

}

resource "google_network_connectivity_service_connection_policy" "service_connection_policies" {
  project       = var.gcp_project_id
  location      = var.gcp_region
  name          = "psc-valkey-${random_id.suffix.hex}"
  service_class = "gcp-memorystore"
  description   = "PSC for Valkey ${random_id.suffix.hex}"
  network       = "projects/${var.gcp_project_id}/global/networks/${google_compute_network.vpc.name}"

  psc_config {
    subnetworks = [
      "projects/${var.gcp_project_id}/regions/${var.gcp_region}/subnetworks/${google_compute_subnetwork.valkey-subnet.name}"
    ]
  }
}
