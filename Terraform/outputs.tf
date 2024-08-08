output "secret_path" {
  value = google_secret_manager_secret.secret.name
}
