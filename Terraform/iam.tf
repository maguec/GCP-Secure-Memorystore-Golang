resource "google_service_account" "service_account" {
  project      = var.gcp_project_id
  account_id   = "account-${random_id.suffix.hex}"
  display_name = "Service Account ${random_id.suffix.hex}"
}

resource "google_project_iam_binding" "secret_viewer" {
  project = var.gcp_project_id
  role    = "roles/secretmanager.viewer"
  members = [
    "serviceAccount:${google_service_account.service_account.email}"
  ]
}

resource "google_project_iam_binding" "secret_accessor" {
  project = var.gcp_project_id
  role    = "roles/secretmanager.secretAccessor"
  members = [
    "serviceAccount:${google_service_account.service_account.email}"
  ]
}


resource "google_project_iam_binding" "memorystore_viewer" {
  project = var.gcp_project_id
  role    = "roles/redis.viewer"
  members = [
    "serviceAccount:${google_service_account.service_account.email}"
  ]
}
