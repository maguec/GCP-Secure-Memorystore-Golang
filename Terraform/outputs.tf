output "secret_path_auth" {
  value = google_secret_manager_secret.secret-auth.name
}

output "secret_path_ip" {
  value = google_secret_manager_secret.secret-ip.name
}

output "secret_path_cert" {
  value = google_secret_manager_secret.secret-cert.name
}

output "vm_ssh_command" {
  value = "gcloud compute ssh --zone ${var.gcp_region}-a vm-${random_id.suffix.hex} --project ${var.gcp_project_id}"
}

output "run_test_command" {
  value =  "/usr/lib/go-1.22/bin/go run secure-pool-example.go --project ${var.gcp_project_id} --instance memorystore-${random_id.suffix.hex}"
}

output "vm_secret_auth" {
  value = "gcloud secrets versions access latest --secret=memorystore-${random_id.suffix.hex}-auth"
}

output "vm_secret_ip" {
  value = "gcloud secrets versions access latest --secret=memorystore-${random_id.suffix.hex}-ip"
}

output "vm_secret_cert" {
  value = "gcloud secrets versions access latest --secret=memorystore-${random_id.suffix.hex}-cert"
}

output "vm_secret_port" {
  value = "gcloud secrets versions access latest --secret=memorystore-${random_id.suffix.hex}-port"
}


