resource "google_memorystore_instance" "valkey_cluster" {
  project        = var.project_id
  instance_id    = var.instance_id
  shard_count    = var.shard_count
  engine_version = var.engine_version
  mode           = var.mode

  desired_psc_auto_connections {
    network    = "projects/${coalesce(var.network_project, var.project_id)}/global/networks/${var.network}"
    project_id = var.project_id
  }

  location                = var.location
  replica_count           = var.replica_count
  node_type               = var.node_type
  transit_encryption_mode = var.transit_encryption_mode
  authorization_mode      = var.authorization_mode
  engine_configs          = var.engine_configs

}