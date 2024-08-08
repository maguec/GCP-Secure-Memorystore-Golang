output "secret_path" {
  value = google_secret_manager_secret.secret.name
}

output "vm_ssh_command" {
  value = "gcloud compute ssh --zone ${var.gcp_region}-a vm-${random_id.suffix.hex} --project ${var.gcp_project_id}"
}

output "vm_secret" {
  value = "gcloud secrets versions access latest --secret=memorystore-${random_id.suffix.hex}"
}

